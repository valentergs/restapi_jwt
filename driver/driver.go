package driver

import (
	"database/sql"
	"log"
)

var db *sql.DB

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//ConnectDB will be exported to main.go
func ConnectDB() *sql.DB {

	var err error

	// Connect to the Postgres Database
	db, err = sql.Open("postgres", "user=rodrigovalente password=password host=localhost port=5432 dbname=api_jwt sslmode=disable")
	logFatal(err)

	return db

}
