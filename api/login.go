package api

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

type LoginManager struct {
	db *sql.DB
}

func NewLoginManager(dbUrl string) *LoginManager {
	db := newDb(dbUrl)
	message := fmt.Sprintf("login.go::NewLoginManager - New LoginManager created with db: %v", dbName)
	log.Default().Println(message)
	return &LoginManager{db: db}
}

// login logs in a user with the given username and password
// It returns a JWT token if the login is successful, or an error if it fails
// This is useful in this context if the user has already registered
// but their last session has expired.
//
// @param sd SignupData - the username and password of the user
// @return string - the JWT token if the login is successful or an error if it fails.
func (lm *LoginManager) login(sd SignupData, doesUserExist func(string) bool) (string, error) {
	if isInvalidData(sd) {
		return "", ErrInvalidData
	}

	if !doesUserExist(sd.Username) {
	    return "", ErrUserDoesNotExist	
	}

	userId, hashedPassword, err := lm.getPassword(sd.Username)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(sd.Password))
	if err != nil {
		return "", ErrWrongPassword
	}

	token, err := signJwt(sd.Username, strconv.Itoa(userId))
	if err != nil {
		return "", err
	}

	message := fmt.Sprintf("login.go::Login - User %s logged in successfully", sd.Username)
	log.Default().Println(message)
	return token, nil
}

func (lm *LoginManager) getPassword(username string) (int, string, error) {
	var userId int
	var password string
	err := lm.db.QueryRow(getPasswordQuery, username).Scan(&userId, &password)
	if err != nil {
		message := fmt.Sprintf("login.go::getPassword - Error querying database: %v", err)
		log.Default().Println(message)
		return 0, "", err
	}

	return userId, password, nil
}
