package api

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func isInvalidData(sd SignupData) bool {
	return sd.Username == "" || sd.Password == "" || len(sd.Username) < 4 || len(sd.Password) < 8
}

func newDb(dbUrl string) *sql.DB {
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("utils::newDb - Error opening database connection: ", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("utils::newDb - Error pinging database: ", err)
	}

	return db
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

func verifyToken(tk string) bool {
	// Parse and verify the JWT token
    token, err := jwt.Parse(tk, func(token *jwt.Token) (interface{}, error) {
        // Check the signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return jwtSecret, nil
    })

    if err != nil {
		message := fmt.Sprintf("Error parsing token: %v", err)
		log.Default().Println(message)
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Default().Printf("Token is valid until %s for user %v\n", claims["exp"], claims["iss"])
        return true
    }
    
	log.Default().Println("Token is invalid")
	return false
}
