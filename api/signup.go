package api

import (
	"errors"
	"fmt"
	"database/sql"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v4"
)

var (
	jwtSecret = []byte("18yr6!b3@3r7")
	db *sql.DB
)

type SignupData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func init() {
	// Initialize the database connection
	var err error
	db, err = sql.Open("postgres", dbName)
	if err != nil {
		log.Fatal("signup.go::init - Error opening database connection: ", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("signup.go::init - Error pinging database: ", err)
	}
}

func signup(sd SignupData) (string, error) {
	// This function would check the database for a duplicate username
	hashedPassword, err := hashPassword(sd.Password)
	if err != nil {
		// logging occurs in the hashPassword function
		return "", err
	}

	if !createNewUser(sd.Username, hashedPassword) {
		// logging occurs in the createNewUser function
		return "", errors.New("Error creating new user")
	}

	token, err := signJwt(sd.Username)
	if err != nil {
		// logging occurs in the signJwt function
		return "", err
	}

	message := fmt.Sprintf("signup.go::signup - User %s signed up successfully", sd.Username)
	log.Default().Println(message)
	return token, nil
}

func doesUserExist(username string) bool {
	rows, err := db.Query(checkUserExistsQuery, username)
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

func hashPassword(password string) (string, error) {
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

func createNewUser(username string, password string) bool {
	// This function would create a new user in the database
	// with the given username and hashed password.
	_, err := db.Exec(createNewUserQuery, username, password)
	if err != nil {
		log.Default().Println("signup.go::createNewUser - Error creating new user: ", err)
		return false
	}

	log.Default().Println("signup.go::createNewUser - New user created successfully")
	return true
}

func signJwt(username string) (string, error) {
	// This function signs a JWT token with the given username
	claims := &jwt.StandardClaims{
        ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
        Issuer:    "go-dropbox",
    }

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        fmt.Println("signup.go::signJwt - Error signing token:", err)
        return "", err
    }

	return tokenString, nil
}