package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

type (
	ProjectData struct {
		Username    string                 `json:"username"`
		Name        string                 `json:"name"`
		Directories map[string]interface{} `json:"directories"`
		Files       []string               `json:"files"`
	}

	ProjectManager struct {
		db *sql.DB
	}
)

var (
	defaultDirectories = map[string]interface{}{
		"root": map[string]interface{}{
			"files": []interface{}{},
		},
	}
)

func (pd ProjectData) String() string {
	return fmt.Sprintf("username: %s, name: %s, directories: %v, files: %v", pd.Username, pd.Name, pd.Directories, pd.Files)

}

func NewProjectManager(dbUrl string) *ProjectManager {
	db := newDb(dbUrl)
	message := fmt.Sprintf("projects.go::NewProjectManager - New ProjectManager created with db: %v", dbUrl)
	log.Default().Println(message)
	return &ProjectManager{db: db}
}

func (pm *ProjectManager) createProject(pd ProjectData, userId int) error {
	if projectExists := pm.doesProjectExist(pd.Name, userId); projectExists {
		log.Default().Println("Project already exists")
		return ErrProjectAlreadyExists
	}

	// Handle the case where the user does not provide any directories
	if pd.Directories == nil {
		pd.Directories = defaultDirectories
	}

	jsonDirectories, err := json.Marshal(pd.Directories)
	if err != nil {
		message := fmt.Sprintf("projects.go::createProject - Error marshalling directories: %v", err)
		log.Default().Println(message)
		return err
	}

	timestamp := time.Now().Unix()
	_, err = pm.db.Exec(createProjectQuery, pd.Name, jsonDirectories, userId, timestamp, timestamp)
	if err != nil {
		message := fmt.Sprintf("projects.go::createProject - Error creating project: %v", err)
		log.Default().Println(message)
		return err
	}

	message := fmt.Sprintf("projects.go::createProject - Project %s created successfully", pd.Name)
	log.Default().Println(message)
	return nil
}

//lint:ignore U1000 Ignore unused function
func (pm *ProjectManager) getUserId(username string) (int, error) {
	var userId int
	err := pm.db.QueryRow(getUserQuery, username).Scan(&userId)
	if err != nil {
		return 0, fmt.Errorf("projects.go::getUserId - Error querying database: %w", err)
	}

	return userId, nil
}

func (pm *ProjectManager) doesProjectExist(name string, userId int) bool {
	txn, err := pm.db.Begin()
	defer func(){
		if err := txn.Commit(); err != nil {
			log.Default().Printf("projects.go::doesProjectExist - Error committing transaction. Err: %v", err)
		}
	}()
	if err != nil {
		log.Default().Printf("projects.go::doesProjectExist - Error beginning transaction. Err: %v", err)
	}
	rows, err := pm.db.Query(checkProjectExistsQuery, name, userId)
	if err != nil {
		message := fmt.Sprintf("projects.go::doesProjectExist - Error querying database: %v", err)
		log.Default().Println(message)
		return false
	}

	return rows.Next()
}

func (pm *ProjectManager) viewProject(projectName, userName string, userId int) ([]byte, error) {
	var projectDirectories []byte
	cachedProjectName := projectName+":"+userName+":"+strconv.Itoa(userId)
	if projectDirectories = redisClient.getDataFromRedisCache(cachedProjectName); projectDirectories != nil {
		return projectDirectories, nil
	}

	err := pm.db.QueryRow(viewProjectQuery, projectName, userId).Scan(&projectDirectories)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProjectDoesNotExist
		}

		message := fmt.Sprintf("projects.go::viewProject - Error querying database: %v", err)
		log.Default().Println(message)
		return nil, err
	}

	log.Default().Printf("projects.go::viewProject - Project %s viewed successfully for user %s\n", projectName, userName)
	redisClient.setDataInRedisCache(cachedProjectName, projectDirectories)
	return projectDirectories, nil
}

func (pm *ProjectManager) deleteProject(projectName string, userId int) error {
	// It should be noted that when a project is deleted, all the files and directories
	// associated with the project are also deleted.
	if !pm.doesProjectExist(projectName, userId) {
		log.Default().Println("Project does not exist")
		return ErrProjectDoesNotExist
	}

	_, err := pm.db.Exec(deleteProjectQuery, projectName, userId)
	if err != nil {
		message := fmt.Sprintf("projects.go::deleteProject - Error deleting project: %v", err)
		log.Default().Println(message)
		return err
	}

	message := fmt.Sprintf("projects.go::deleteProject - Project %s deleted successfully", projectName)
	log.Default().Println(message)
	return nil
}
