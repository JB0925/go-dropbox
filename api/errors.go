package api

import "errors"

var (
	ErrProjectDoesNotExist = errors.New("project does not exist")
	ErrorFileAlreadyExists = errors.New("file already exists")
	ErrFileDoesNotExist        = errors.New("file does not exist")
    ErrUnauthorized  = errors.New("user does not have access to file/project")
	ErrMissingRequiredFields = errors.New("missing required fields")
	ErrUserDoesNotExist = errors.New("user does not exist")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidData = errors.New("invalid data")
	ErrWrongPassword = errors.New("wrong password")
	ErrInvalidPath = errors.New("invalid path")
	ErrFileChanged = errors.New("file changed since it was shared")
)