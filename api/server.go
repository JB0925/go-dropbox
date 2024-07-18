package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
)

const (
	dbName = "postgresql:///dropbox?sslmode=disable"
)

func NewServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/signup", signupUser)

	return &http.Server{
		Handler: mux,
		Addr:   ":8080",
	}
}

func StartServer(s *http.Server) error {
	// Try to start the server and run it in a goroutine so that multiple requests can be handled concurrently
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
		message := fmt.Sprintf("signup.go::signup - Error decoding request body: %v", err)
		log.Default().Println(message)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if isInvalidData(sd) {
		message := fmt.Sprintf("signup.go::signup - Invalid data - Username: %s, Password: %s", sd.Username, sd.Password)
		log.Default().Println(message)
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if doesUserExist(sd.Username) {
		message := fmt.Sprintf("signup.go::signup - User with username %s already exists", sd.Username)
		log.Default().Println(message)
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	token, err := signup(sd)
	if err != nil {
		message := fmt.Sprintf("signup.go::signup - Error signing up: %v", err)
		log.Default().Println(message)
		http.Error(w, "Error signing up", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
