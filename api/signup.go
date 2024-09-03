package api

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecret = []byte("18yr6!b3@3r7")
)

type SignupData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignupManager struct {
	db *sql.DB
}

func NewSignupManager(dbUrl string) *SignupManager {
	db := newDb(dbUrl)
	message := fmt.Sprintf("signup.go::NewSignupManager - New SignupManager created with db: %v", dbUrl)
	log.Default().Println(message)
	return &SignupManager{db: db}
}

func (sm *SignupManager) signup(sd SignupData) (string, error) {
	// This function would check the database for a duplicate username
	hashedPassword, err := sm.hashPassword(sd.Password)
	if err != nil {
		// logging occurs in the hashPassword function
		return "", err
	}

	userId, err := sm.createNewUser(sd.Username, hashedPassword)
	if err != nil || userId == 0 {
		// logging occurs in the createNewUser function
		return "", errors.New("Error creating new user")
	}

	token, err := signJwt(sd.Username, string(userId))
	if err != nil {
		// logging occurs in the signJwt function
		return "", err
	}

	message := fmt.Sprintf("signup.go::signup - User %s signed up successfully", sd.Username)
	log.Default().Println(message)
	return token, nil
}

func (sm *SignupManager) doesUserExist(username string) bool {
	rows, err := sm.db.Query(checkUserExistsQuery, username)
	if err != nil {
		log.Default().Println("signup.go::doesUserExist - Error checking if user exists: ", err)
		return true
	}

	defer rows.Close()

	if rows.Next() {
		message := fmt.Sprintf("signup.go::doesUserExist - User with username %s already exists", username)
		log.Default().Println(message)
		return true
	}

	message := fmt.Sprintf("signup.go::doesUserExist - User with username %s does not exist", username)
	log.Default().Println(message)
	return false
}

func (sm *SignupManager) hashPassword(password string) (string, error) {
	// A check is done earlier to ensure that the password is not empty.
	// This function would hash the password and return the hashed password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Default().Println("signup.go::hashPassword - Error hashing password: ", err)
		return "", err
	}

	log.Default().Println("signup.go::hashPassword - Password hashed successfully")
	return string(hashedPassword), nil
}

func (sm *SignupManager) createNewUser(username string, password string) (int, error) {
	// This function would create a new user in the database
	// with the given username and hashed password.
	var userId int
	err := sm.db.QueryRow(createNewUserQuery, username, password).Scan(&userId)
	if err != nil {
		log.Default().Println("signup.go::createNewUser - Error creating new user: ", err)
		return 0, errors.New("Error creating new user")
	}

	log.Default().Println("signup.go::createNewUser - New user created successfully")
	return userId, nil
}
