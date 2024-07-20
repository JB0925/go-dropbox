package api

import (
	"database/sql"
	"log"
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
