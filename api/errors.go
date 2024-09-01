package api

import "errors"

var (
	ErrProjectDoesNotExist = errors.New("Project does not exist")
	ErrorFileAlreadyExists = errors.New("File already exists")
	ErrFileDoesNotExist        = errors.New("file does not exist")
    ErrUnauthorized  = errors.New("user does not have access to file/project")
	ErrMissingRequiredFields = errors.New("Missing required fields")
	ErrUserDoesNotExist = errors.New("User does not exist")
	ErrUserAlreadyExists = errors.New("User already exists")
	ErrInvalidData = errors.New("Invalid data")
	ErrWrongPassword = errors.New("Wrong password")
	ErrInvalidPath = errors.New("Invalid path")
)