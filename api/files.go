package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type (
	FileData struct {
		Name        string `json:"name"`
		Path        string `json:"path"`
		ProjectName string `json:"project_name"`
	}

	SharingData struct {
		Name        string `json:"name"`
		ProjectName string `json:"project_name"`
		UserId      int    `json:"user_id"`
		ProjectId   int    `json:"project_id"`
	}

	FileManager struct {
		db *sql.DB
	}
)

var (
	fileSharingUserId = 1
)

func NewFileManager(dbUrl string) *FileManager {
	db := newDb(dbUrl)
	message := fmt.Sprintf("files.go::NewFileManager - New FileManager created with db: %v", dbUrl)
	log.Default().Println(message)
	return &FileManager{db: db}
}

func (fm FileManager) upload(fd FileData, username string, fileContent []byte) error {
	// This function checks the database for a duplicate file name
	// and then upload the file to the database.
	// It returns an error if the file already exists
	// or if there was an error uploading the file.
	//
	// @param: fd FileData - the file data to be uploaded, taken from the request body
	// @param: username string - the username of the user uploading the file
	// @param: fileContent []byte - the content of the file to be uploaded
	// @return: An error if one exists
	projectId, userId, dirs, err := fm.getFileOwnerData(fd.ProjectName, username)
	if err != nil {
		return err
	}

	exists, err := fm.doesFileExist(projectId, userId, fd.Name, fd.Path)
	if err != nil {
		return err
	}

	if exists {
		log.Default().Println("files.go::upload - File already exists")
		return ErrorFileAlreadyExists
	}

	directories, err := fm.parseDirectories(dirs)
	if err != nil {
		message := fmt.Sprintf("files.go::upload - Error parsing directories: %v", err)
		log.Default().Println(message)
		return err
	}

	err = fm.findAndInsertPath(directories, fd.Path, fd.Name)
	if err != nil {
		message := fmt.Sprintf("files.go::upload - Error finding and inserting path: %v", err)
		log.Default().Println(message)
		return err
	}

	log.Default().Println("files.go::upload - Directories: ", directories)

	timestamp := time.Now().Unix()
	err = fm.storeFile(fd, fileContent, projectId, userId, timestamp)
	if err != nil {
		message := fmt.Sprintf("files.go::upload - Error storing file: %v", err)
		log.Default().Println(message)
		return err
	}

	err = fm.updateProjectStructure(projectId, directories, timestamp)
	if err != nil {
		message := fmt.Sprintf("files.go::upload - Error updating project structure: %v", err)
		log.Default().Println(message)
		return err
	}

	return nil
}

func (fm FileManager) download(projectName, fileName string, userId int) ([]byte, error) {
	// This function gets the file content from the database
	// and returns an error if the file does not exist.
	//
	// @param: projectName string - the name of the project
	// @param: fileName string - the name of the file
	// @return: []byte - the content of the file
	// @return: error - An error if one exists
	cachedFileName := projectName+":"+fileName
	if userId != fileSharingUserId {
		d, err := redisClient.redisClient.Get(ctx, cachedFileName).Result()
		if err != nil {
			log.Default().Printf("files.go::upload - redis error on get %s: %v", cachedFileName, err)
		}

		if d != "" {
			log.Default().Printf("files.go::upload - found %s in redis cache with len %d", cachedFileName, len([]byte(d)))
			return []byte(d), nil
		}
	}

	var data []byte
	var fileUserId int
	err := fm.db.QueryRow(getFileQuery, fileName, projectName).Scan(&data, &fileUserId)
	if err != nil {
		message := fmt.Sprintf("files.go::get - Error querying database: %v", err)
		log.Default().Println(message)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFileDoesNotExist
		}
		return nil, err
	}

	if len(data) == 0 {
		return nil, ErrFileDoesNotExist
	}

	if userId != fileSharingUserId && fileUserId != userId {
		message := fmt.Sprintf("files.go::get - User with id %d does not have access file %s", userId, fileName)
		log.Default().Println(message)
		return nil, ErrUnauthorized
	}

	if userId != fileSharingUserId {
		err = redisClient.redisClient.Set(ctx, cachedFileName, data, time.Duration(24 * time.Hour)).Err()
		if err != nil {
			log.Default().Printf("files.go::upload - redis error on set %s: %v", cachedFileName, err)
		}
	}
	return data, nil
}

func (fm FileManager) getProjectIdAndDirectories(projectName string, userId int) (int, []byte, error) {
	// This function gets the project id and directories from the database
	// and returns an error if the project does not exist.
	//
	// @param: projectName string - the name of the project
	// @return: int - the project id
	// @return: []byte - the directories of the project
	// @return: error - An error if one exists
	rows, err := fm.db.Query(getProjectIdQuery, projectName, userId)
	if err != nil {
		message := fmt.Sprintf("files.go::getProjectId - Error querying database: %v", err)
		log.Default().Println(message)
		return 0, []byte{}, err
	}

	defer rows.Close()

	var projectId int
	var directories []byte
	if rows.Next() {
		err = rows.Scan(&projectId, &directories)
		if err != nil {
			message := fmt.Sprintf("files.go::getProjectId - Error scanning database: %v", err)
			log.Default().Println(message)
			return 0, []byte{}, err
		}

		return projectId, directories, nil
	}

	return 0, []byte{}, ErrProjectDoesNotExist
}

func (fm *FileManager) getUserId(username string) (int, error) {
	// Given a username, this function queries the database
	// and gets the user id of the associated username if it exists.
	// It returns an error if the user does not exist.
	//
	// @param: username string - the username of the user
	// @return: int - the user id
	// @return: error - An error if one exists
	var userId int
	err := fm.db.QueryRow(getUserQuery, username).Scan(&userId)
	if err != nil {
		message := fmt.Sprintf("files.go::getUserId - Error querying database: %v", err)
		log.Default().Println(message)
		return 0, err
	}

	return userId, nil
}

func (fm *FileManager) doesFileExist(project_id, user_id int, fileName, filePath string) (bool, error) {
	// This function checks the database for the file
	// and return true if the file exists
	// and false if the file does not exist.
	// It also returns an error - nil if no error exists, the error if one does exist.
	//
	// @param: project_id int - the project id of the file
	// @param: user_id int - the user id of the file
	// @return: bool - true if the file exists, false if it does not
	// @return: error - An error if one exists
	rows, err := fm.db.Query(checkFileExistsQuery, project_id, user_id, fileName, filePath)
	if err != nil {
		message := fmt.Sprintf("files.go::doesFileExist - Error querying database: %v", err)
		log.Default().Println(message)
		return false, err
	}

	defer rows.Close()

	if rows.Next() {
		return true, nil
	}

	return false, nil
}

//lint:ignore U1000 Ignore unused function in case of potential future use
func (fm *FileManager) getDirectories(projectId int) ([]byte, error) {
	// *** DEPRECATED ***
	// Gets the directories map associated with the project.
	// Is not needed as the directories are taken as a part of another query,
	// thus avoiding an extra call to the database.
	var directories []byte
	err := fm.db.QueryRow(getProjectDirectoriesQuery, projectId).Scan(&directories)
	if err != nil {
		message := fmt.Sprintf("files.go::getDirectories - Error querying database: %v", err)
		log.Default().Println(message)
		return nil, err
	}

	return directories, nil
}

func (fm *FileManager) parseDirectories(directories []byte) (map[string]interface{}, error) {
	// This function parses the directories from the database
	// and turns them into JSON from a byte slice.
	//
	// @param: directories []byte - the directories from the database
	// @return: map[string]interface{} - the parsed directories
	// @return: error - An error if one exists
	var dirs map[string]interface{}
	err := json.Unmarshal(directories, &dirs)

	if err != nil {
		message := fmt.Sprintf("files.go::parseDirectories - Error unmarshalling directories: %v", err)
		log.Default().Println(message)
		return nil, err
	}

	return dirs, nil
}

func (fm *FileManager) findAndInsertPath(directories map[string]interface{}, filePath, fileName string) error {
	// findAndInsertPath takes a map[string]interface{}, a filePath, and a fileName
	// The function finds <filePath> within <directories> ( or creates it if it does not exist )
	// It then adds <fileName> into the files array in that directory.
	//
	// @param directories - map[string]interface{}: a nested map of directories
	//     Ex: {"root": {"files": [], "bar": {"files": []}}}}'}
	//         - bar is a nested directory within the root of the project
	//
	// @param filePath - string: the file path at which to insert the new file name
	// @param fileName - string: the new file name to insert
	// @return error - error: An error if one occurred, nil otherwise
	segments := strings.Split(strings.Trim(filePath, "/"), "/")
	if segments[0] != "root" {
		return ErrInvalidPath
	}

	// Traverse the directories map to find the correct path
	current := directories

	// Loop over each directory in the file path that the user gave
	for _, segment := range segments {
		if _, exists := current[segment]; !exists {
			// Create the new directory if it does not exist
			current[segment] = map[string]interface{}{
				"files": []interface{}{},
			}
		}

		// Move to the next directory
		current = current[segment].(map[string]interface{})
	}

	// Append the file to the "files" array - or create a new one if it does not exist
	if files, ok := current["files"].([]interface{}); ok {
		current["files"] = append(files, fileName)
	} else {
		current["files"] = []interface{}{fileName}
	}

	return nil
}

func (fm *FileManager) updateProjectStructure(
	// This function marshals the updated directory structure
	// and updates it in the database.
	//
	// @param projectId - int: the id of the project
	// @param directories - map[string]interface{} - the directories that make up the project
	// @param timestamp - int64: a Unix timestamp that represents when the project was updated
	// @return error - error: An error if one occurred, nil otherwise
	projectId int,
	directories map[string]interface{},
	timestamp int64) error {
	// This function updates the project structure in the database
	directoriesToJson, err := json.Marshal(directories)
	if err != nil {
		message := fmt.Sprintf("files.go::updateProjectStructure - Error marshalling directories: %v", err)
		log.Default().Println(message)
		return err
	}

	_, err = fm.db.Exec(updateProjectDirectoryQuery, directoriesToJson, projectId, timestamp)
	if err != nil {
		message := fmt.Sprintf("files.go::updateProjectStructure - Error updating project structure: %v", err)
		log.Default().Println(message)
		return err
	}

	return nil
}

func (fm *FileManager) getFileOwnerData(projectName, username string) (int, int, []byte, error) {
	// One function to get several aspects of user and project related data,
	// such as the project id, user id, and directories.
	//
	// @param: projectName string - the name of the project
	// @param: username string - the username of the user
	// @return: int - the project id
	// @return: int - the user id
	// @return: []byte - the directories of the project
	// @return: error - An error if one exists
	userId, err := fm.getUserId(username)
	if err != nil {
		return 0, 0, []byte{}, err
	}

	projectId, directories, err := fm.getProjectIdAndDirectories(projectName, userId)
	if err != nil {
		return 0, 0, []byte{}, err
	}

	return projectId, userId, directories, nil
}

func (fm *FileManager) storeFile(
	fd FileData,
	fileContent []byte,
	projectId,
	userId int,
	timestamp int64) error {
	// A wrapper method used to call a database and store the contents of a file
	// and its related metadata.
	//
	// @param: fd FileData - the file data to be uploaded
	// @param: fileContent []byte - the content of the file to be uploaded
	// @param: projectId int - the project id of the file
	// @param: userId int - the user id of the file
	// @return: error - An error if one exists
	_, err := fm.db.Exec(
		uploadFileQuery,
		fd.Name,
		fd.Path,
		fileContent,
		userId,
		projectId,
		timestamp,
		timestamp)

	if err != nil {
		message := fmt.Sprintf("files.go::upload - Error uploading file: %v", err)
		log.Default().Println(message)
		return err
	}

	return nil
}

func (fm *FileManager) findAndDeleteFileFromDirectories(
	// Finds the file within the directory structure and deletes it.
	//
	// @param fileName - string: the name of the file
	// @param filePath - string: the path from which to delete the file
	// @param directories - map[string]interface{}: the directories that represent the project structure
	// @return error - error: An error if one occurred, nil otherwise
	fileName,
	filePath string,
	directories map[string]interface{}) error {
	segments := strings.Split(strings.Trim(filePath, "/"), "/")
	log.Default().Println("files.go::delete - Segments: ", segments[0])
	if segments[0] != "root" {
		return ErrInvalidPath
	}

	// Traverse the directories map to find the correct path
	current := directories
	log.Default().Println("files.go::delete - Current at Start: ", current)

	// Loop over each directory in the file path that the user gave
	for _, segment := range segments {
		if _, exists := current[segment]; !exists {
			// Return file not found if the current part of the path does not exist
			log.Default().Println("files.go::delete - File does not exist. Current path segment: ", segment)
			return ErrFileDoesNotExist
		}

		// Move to the next directory
		current = current[segment].(map[string]interface{})
	}

	filesInCurrentDir, ok := current["files"].([]interface{})
	if !ok {
		// if the "files" array does not exist, return file not found
		// note that this should not happen - every stored directory should have a "files" array
		return ErrFileDoesNotExist
	}

	// Remove the file from the "files" array
	fm.removeFileFromFilesArray(filesInCurrentDir, fileName, current)
	log.Default().Println("files.go::delete - Directories: ", directories)
	return nil
}

func (fm *FileManager) deleteFile(projectName, fileName, filePath, userName string) error {
	// This function deletes a file from the database
	// and returns an error if the file does not exist.
	//
	// @param: projectName string - the name of the project
	// @param: fileName string - the name of the file
	// @param: filePath string - the path of the file
	// @param: userName string - the username of the user
	// @return: error - An error if one exists
	projectId, userId, dirs, err := fm.getFileOwnerData(projectName, userName)
	if err != nil {
		return err
	}

	dirsMap, err := fm.parseDirectories(dirs)
	if err != nil {
		message := fmt.Sprintf("files.go::delete - Error parsing directories: %v", err)
		log.Default().Println(message)
		return err
	}

	err = fm.findAndDeleteFileFromDirectories(fileName, filePath, dirsMap)
	if err != nil {
		message := fmt.Sprintf("files.go::delete - Error finding and deleting file from directories: %v", err)
		log.Default().Println(message)
		return err
	}

	err = fm.removeFileFromDataStore(fileName, filePath, projectName, projectId, userId)
	if err != nil {
		message := fmt.Sprintf("files.go::delete - Error removing file from data store: %v", err)
		log.Default().Println(message)
		return err
	}

	err = fm.updateProjectStructure(projectId, dirsMap, time.Now().Unix())
	if err != nil {
		message := fmt.Sprintf("files.go::delete - Error updating project structure: %v", err)
		log.Default().Println(message)
		return err
	}

	return nil
}

func (fm *FileManager) removeFileFromFilesArray(
	// This function removes a file from the "files" array in the directories map
	// And updates the directories map with the new "files" array - both in place
	//
	// @param: filesInCurrentDir []interface{} - the "files" array in the current directory
	// @param: fileName string - the name of the file to be removed
	// @param: current map[string]interface{} - the current directory
	filesInCurrentDir []interface{},
	fileName string,
	current map[string]interface{}) {
	for i, file := range filesInCurrentDir {
		if file == fileName {
			filesInCurrentDir = append(filesInCurrentDir[:i], filesInCurrentDir[i+1:]...)
			current["files"] = filesInCurrentDir
			break
		}
	}
}

func (fm *FileManager) removeFileFromDataStore(
	// This function removes a file from the database
	// and returns an error if the file does not exist.
	//
	// @param: fileName string - the name of the file
	// @param: projectId int - the project id of the file
	// @param: userId int - the user id of the file
	fileName,
	filePath,
	projectName string,
	projectId,
	userId int) error {
	_, err := fm.db.Exec(deleteFileQuery, fileName, projectName, userId, projectId, filePath)
	if err != nil {
		message := fmt.Sprintf("files.go::delete - Error deleting file: %v", err)
		log.Default().Println(message)
		return err
	}

	log.Default().Printf("files.go::delete - File %s deleted successfully", fileName)
	return nil
}

func (fm *FileManager) update(
	fd FileData,
	fileContent []byte,
	userId int) error {
	// This function updates a file in the database
	// and returns an error if the file does not exist or
	// if there was an error updating the file.
	//
	// @param fd - FileData: a struct that contains the file name, path, and project name
	// @param fileContent - []byte: A byte array containing the contents of the updated file
	// @param userId - int: an integer representing the user ID
	// @return error = error: An error if one occurred, nil otherwise
	projectId, _, err := fm.getProjectIdAndDirectories(fd.ProjectName, userId)
	if err != nil {
		msg := fmt.Errorf("files.go::update - Error getting project id with project name %s. Error: %w", fd.ProjectName, err)
		log.Default().Println(msg)
		return err
	}

	exists, err := fm.doesFileExist(projectId, userId, fd.Name, fd.Path)
	if err != nil {
		log.Default().Println("files.go::update - Error checking if file exists: ", err)
		return err
	}

	if !exists {
		log.Default().Println("files.go::update - File does not exist")
		return ErrFileDoesNotExist
	}

	timestamp := time.Now().Unix()
	_, err = fm.db.Query(updateFileQuery, fileContent, timestamp, fd.Name, projectId, userId)
	if err != nil {
		message := fmt.Sprintf("files.go::update - Error updating file: %v", err)
		log.Default().Println(message)
		return err
	}

	return nil
}

func (fm *FileManager) createHashFromContent(fileContent []byte) (string, error) {
	// This function takes in the content of a file and hashes the bytes.
	// This is used later to tell if the file changed, which means it should not be shared.
	//
	// @param fileContent - byte[]: A byte array containing the contents of the updated file
	// @return string: A string representing the hash of the file's contents
	// @return error: An error if one occurred, nil otherwise
	h := sha256.New()
	if _, err := h.Write(fileContent); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (fm *FileManager) storeFileHashForSharing(sd SharingData) (string, error) {
	// After the file's content is hashed, this function stores the hash in the shared table
	// 
	// @param sd - SharingData: A struct containing the file name, project name, user ID, and project ID
	// @return string: The hash that was stored.
	// @return error: An error if one occurred, nil otherwise.
	var uid int
	var fileData []byte
	if err := fm.db.QueryRow(getFileQuery, sd.Name, sd.ProjectName).Scan(&fileData, &uid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrFileDoesNotExist
		}
		return "", fmt.Errorf("error getting file. Err: %w", err)
	}
	if sd.UserId != uid {
		return "", fmt.Errorf("error storing hash for shared file. User ids do not match")
	}
	hash, err := fm.createHashFromContent(fileData)
	if err != nil {
		return "", fmt.Errorf("error creating hash of file content. Err: %w", err)
	}
	_, err = fm.db.Exec(createSharedFileQuery, hash, sd.UserId, sd.ProjectId)
	if err != nil {
		return "", fmt.Errorf("error storing hash for file sharing. Err: %w", err)
	}

	return hash, nil
}

func (fm *FileManager) shareFile(userGivenHash string) (string, []byte, error) {
	// This function is used to check if a file hash exists in the "shared" table and, 
	// if so, downloads it for the user who wants it.
	//
	// @param userGivenHash - string: The hash supplied by the user that wants the shared file.
	// @param userName - string: The name of the user who shared the file.
	// @return string: the name of the file
	// @return []byte: the contents of the file with the given file name
	// @return error: An error if one occurred, nil otherwise
	var projectName string
	var fileName string

	if err := fm.db.QueryRow(
		getSharedFileDetailsQuery, userGivenHash,
	).Scan(&fileName, &projectName); err != nil {
		errMsg := fmt.Errorf("an error occurred when trying to find the given hash for sharing. Err: %w", err)
		log.Default().Println(errMsg.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return "", []byte{}, ErrFileDoesNotExist
		}
		return "", []byte{}, fmt.Errorf("error getting shared file: %w", err)
	}

	// at this point, we've determined that this file exists as expected,
	// so now we can call "download" to return it to the user.
	fileData, err := fm.download(projectName, fileName, fileSharingUserId)
	if err != nil {
		return "", []byte{}, fmt.Errorf("could not download file %s. Err: %w", fileName, err)
	}
	currentHash, err := fm.createHashFromContent(fileData)
	if err != nil {
		return "", []byte{}, fmt.Errorf("could not generate hash of file content. Err: %w", err)
	}
	if currentHash != userGivenHash {
		return "", []byte{}, ErrFileChanged
	}
	return fileName, fileData, err
}
