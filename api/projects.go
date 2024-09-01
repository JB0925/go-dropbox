package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type ProjectData struct {
	Username    string `json:"username"`
	Name        string `json:"name"`
	Directories	map[string]interface{} `json:"directories"`
	Files	    []string `json:"files"`
}

type ProjectManager struct {
	db *sql.DB
}

func (pd ProjectData) String() string {
	return fmt.Sprintf("username: %s, name: %s, directories: %v, files: %v", pd.Username, pd.Name, pd.Directories, pd.Files)

}

func NewProjectManager(dbUrl string) *ProjectManager {
	db := newDb(dbUrl)
	message := fmt.Sprintf("projects.go::NewProjectManager - New ProjectManager created with db: %v", dbUrl)
	log.Default().Println(message)
	return &ProjectManager{db: db}
}

func (pm *ProjectManager) createProject(pd ProjectData) error {
	userId, err := pm.getUserId(pd.Username)
	if err != nil || userId == 0 {
		return err
	}

	if projectExists := pm.doesProjectExist(pd.Name); projectExists {
		log.Default().Println("Project already exists")
		return ErrProjectAlreadyExists
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

func (pm *ProjectManager) getUserId(username string) (int, error) {
	var userId int
	err := pm.db.QueryRow(getUserQuery, username).Scan(&userId)
	if err != nil {
		message := fmt.Sprintf("projects.go::getUserId - Error querying database: %v", err)
		log.Default().Println(message)
		return 0, err
	}

	return userId, nil
}

func (pm *ProjectManager) doesProjectExist(name string) bool {
	rows, err := pm.db.Query(checkProjectExistsQuery, name)
	if err != nil {
		message := fmt.Sprintf("projects.go::doesProjectExist - Error querying database: %v", err)
		log.Default().Println(message)
		return false
	}

	return rows.Next()
}