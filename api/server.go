package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/jbrink/go-dropbox/rate_limiter"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		err = godotenv.Load("../.env")
		if err != nil {
			 log.Default().Fatal("Error loading .env file")
		}
	}

	dbName = os.Getenv("DATABASE_URL")
	loginManager = NewLoginManager(dbName)
	signupManager = NewSignupManager(dbName)
	projectManager = NewProjectManager(dbName)
	fileManager = NewFileManager(dbName)
}

var (
	dbName string
	loginManager *LoginManager
	signupManager *SignupManager
	projectManager *ProjectManager
	fileManager *FileManager
	ErrProjectAlreadyExists = errors.New("Project already exists")
	rateLimiter = rate_limiter.NewRateLimiter(5, 30 * time.Second, 2)
)

func NewServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/signup", rateLimiter.RateLimit(http.HandlerFunc(signupUser)))
	mux.HandleFunc("/login", rateLimiter.RateLimit(http.HandlerFunc(loginUser)))
	mux.HandleFunc("/projects/create", rateLimiter.RateLimit(checkAuth(http.HandlerFunc(createProject))))
	mux.HandleFunc("/projects/view", rateLimiter.RateLimit(checkAuth(http.HandlerFunc(viewProject))))
	mux.HandleFunc("/projects/delete", rateLimiter.RateLimit(checkAuth(http.HandlerFunc(deleteProject))))
	mux.HandleFunc("/files/upload", rateLimiter.RateLimit(checkAuth(http.HandlerFunc(uploadFile))))
	mux.HandleFunc("/files/download", rateLimiter.RateLimit(checkAuth(http.HandlerFunc(downloadFile))))
	mux.HandleFunc("/files/delete", rateLimiter.RateLimit(checkAuth(http.HandlerFunc(deleteFile))))

	return &http.Server{
		Handler: mux,
		Addr:   ":8080",
	}
}

func StartServer(s *http.Server) error {
	// Try to start the server and run it in a goroutine so that multiple requests can be handled concurrently
	go rateLimiter.Refresh()  // Start the rate limiter refresh in a goroutine

	var e error
	go func() {
		log.Default().Println("server::Start - Starting server on port 8080")
		err := s.ListenAndServe()
		if err != nil {
			log.Default().Printf("server::Start - Error starting server: %v", err)
			e = err
		}
	}()

	// Listen for signals to shut down the server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan  // the server will block here until a signal is received, which will prevent the main process from exiting
	log.Default().Print("server::Start - Shutting down server")
	return e
}

func signupUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var sd SignupData
	err := json.NewDecoder(r.Body).Decode(&sd)
	if err != nil {
		message := fmt.Sprintf("server.go::signupUser - Error decoding request body: %v", err)
		log.Default().Println(message)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if isInvalidData(sd) {
		message := fmt.Sprintf("server.go::signupUser - Invalid data - Username: %s, Password: %s", sd.Username, sd.Password)
		log.Default().Println(message)
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if signupManager.doesUserExist(sd.Username) {
		message := fmt.Sprintf("server.go::signupUser - User with username %s already exists", sd.Username)
		log.Default().Println(message)
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	token, err := signupManager.signup(sd)
	if err != nil {
		message := fmt.Sprintf("server.go::signupUser - Error signing up: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error signing up", getErrorCode(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var sd SignupData
	err := json.NewDecoder(r.Body).Decode(&sd)
	if err != nil {
		message := fmt.Sprintf("server.go::loginUser - Error decoding request body: %v", err)
		log.Default().Println(message)
		http.Error(w, "Invalid request body", getErrorCode(err))
		return
	}

	token, err := loginManager.login(sd, signupManager.doesUserExist)
	if err != nil {
		message := fmt.Sprintf("server.go::loginUser - Error logging in: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error logging in", getErrorCode(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func createProject(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var pd ProjectData
	err := json.NewDecoder(r.Body).Decode(&pd)
	if err != nil {
		message := fmt.Sprintf("server.go::createProject - Error decoding request body: %v", err)
		log.Default().Println(message)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userId, err := getAndConvertUserId(r) // get the user id from the request X-GO-DROPBOX-USER-ID header
	if err != nil {
		message := fmt.Sprintf("server.go::createProject - Error converting userId to int: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error creating project", getErrorCode(err))
		return
	}

	if err = projectManager.createProject(pd, userId); err != nil {
		message := fmt.Sprintf("server.go::createProject - Error creating project: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error creating project", getErrorCode(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func viewProject(w http.ResponseWriter, r *http.Request) {
	projectName := r.URL.Query().Get("project_name")
	if projectName == "" {
		message := fmt.Sprintf("server.go::viewProject - Missing required field: projectName = %s", projectName)
		log.Default().Println(message)
		http.Error(w, "Missing required field", getErrorCode(ErrMissingRequiredFields))
		return
	}

	userName := r.Header.Get("X-GO-DROPBOX-USER")
	userId, err := getAndConvertUserId(r) // get the user id from the request X-GO-DROPBOX-USER-ID header
	if err != nil {
		message := fmt.Sprintf("server.go::viewProject - Error converting userId to int: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error viewing project", getErrorCode(err))
		return
	}

	project, err := projectManager.viewProject(projectName, userName, userId)
	if err != nil {
		message := fmt.Sprintf("server.go::viewProject - Error viewing project: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error viewing project", getErrorCode(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"project": string(project)})
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	err := r.ParseMultipartForm(10 << 20) // 10MB
    if err != nil {
        http.Error(w, "Unable to parse form", http.StatusBadRequest)
        return
    }

	userName := r.Header.Get("X-GO-DROPBOX-USER")
	log.Default().Printf("server.go::uploadFile - User %s is uploading a file\n", userName)

	path := r.FormValue("path")
	name := r.FormValue("name")
	projectName := r.FormValue("project_name")

	if userName == "" || projectName == "" || name == "" || path == "" {
		message := fmt.Sprintf("server.go::uploadFile - Missing required fields: %s, %s, %s, %s", userName, projectName, name, path)
		log.Default().Println(message)
		http.Error(w, "Missing required fields", getErrorCode(ErrMissingRequiredFields))
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		message := fmt.Sprintf("server.go::uploadFile - Error getting file: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error getting file", getErrorCode(err))
		return
	}

	defer file.Close()

	fc, err := io.ReadAll(file)
	if err != nil {
		message := fmt.Sprintf("server.go::uploadFile - Error reading file: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error reading file", getErrorCode(err))
		return
	}

	log.Default().Printf("Got file %s of content length %d\n", name, len(fc))

	fd := FileData{
		Name: name,
		Path: path,
		ProjectName: projectName,
	}

	if err = fileManager.upload(fd, userName, fc); err != nil {
		message := fmt.Sprintf("server.go::uploadFile - Error uploading file: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error uploading file", getErrorCode(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("name")
	projectName := r.URL.Query().Get("project_name")

	if fileName == "" || projectName == "" {
		message := fmt.Sprintf("server.go::downloadFile - Missing required fields: fileName =  %s, projectName = %s", fileName, projectName)
		log.Default().Println(message)
		http.Error(w, "Missing required fields", getErrorCode(ErrMissingRequiredFields))
		return
	}

	message := fmt.Sprintf("server.go::downloadFile - Getting file %s from project %s", fileName, projectName)
	log.Default().Println(message)

	userName := r.Header.Get("X-GO-DROPBOX-USER")
	file, err := fileManager.download(projectName, fileName, userName)
	if err != nil {
		message := fmt.Sprintf("server.go::downloadFile - Error getting file: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error getting file", getErrorCode(err))
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(file)))
	w.WriteHeader(http.StatusOK)
	w.Write(file)	
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var d map[string]string

	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		message := fmt.Sprintf("server.go::deleteFile - Error decoding request body: %v", err)
		log.Default().Println(message)
		http.Error(w, "Invalid request body", getErrorCode(err))
		return
	}

	log.Default().Println("server.go::deleteFile - Got data for delete request. Data: ", d)

	username := r.Header.Get("X-GO-DROPBOX-USER")
	err = fileManager.deleteFile(d["project_name"], d["name"], d["path"], username)
	if err != nil {
		message := fmt.Sprintf("server.go::deleteFile - Error deleting file: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error deleting file", getErrorCode(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	projectName := r.URL.Query().Get("project_name")
	if projectName == "" {
		message := fmt.Sprintf("server.go::deleteProject - Missing required field: projectName = %s", projectName)
		log.Default().Println(message)
		http.Error(w, "Missing required field", getErrorCode(ErrMissingRequiredFields))
		return
	}

	userId, err := getAndConvertUserId(r) // get the user id from the request X-GO-DROPBOX-USER-ID header
	if err != nil {
		message := fmt.Sprintf("server.go::deleteProject - Error converting userId to int: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error deleting project", getErrorCode(err))
		return
	}

	err = projectManager.deleteProject(projectName, userId)
	if err != nil {
		message := fmt.Sprintf("server.go::deleteProject - Error deleting project: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error deleting project", getErrorCode(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
