package api

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
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
	claims := &jwt.StandardClaims{
        ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
        Issuer:    username,
    }

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        fmt.Println("signup.go::signJwt - Error signing token:", err)
        return "", err
    }

	return tokenString, nil
}

func verifyToken(tk string) (bool, string) {
	// Parse and verify the JWT token
	if tk == "" {
		log.Default().Println("No token provided")
		return false, ""
	}

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
        return true, claims["iss"].(string)
    }
    
	log.Default().Println("Token is invalid")
	return false, ""
}

func checkAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		validToken, username := verifyToken(auth)
		if auth == "" || !validToken {
			message := fmt.Sprintf("utils::checkAuth - No authorization header provided by user %s", username)
			log.Default().Println(message)
			http.Error(w, "No authorization header provided or auth is invalid", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-GO-DROPBOX-USER", username)
		next.ServeHTTP(w, r)
	}
}

func getErrorCode(err error) int {
	switch {
	case errors.Is(err, ErrorFileAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, ErrFileDoesNotExist):
		return http.StatusNotFound
	case errors.Is(err, ErrProjectDoesNotExist):
		return http.StatusNotFound
	case errors.Is(err, ErrProjectAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, ErrMissingRequiredFields):
		return http.StatusBadRequest
	case errors.Is(err, ErrUserDoesNotExist):
		return http.StatusUnauthorized
	case errors.Is(err, ErrInvalidData):
		return http.StatusBadRequest
	case errors.Is(err, ErrWrongPassword):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
