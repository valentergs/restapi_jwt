package driver

import (
	"database/sql"
)

var db *sql.DB

//ConnectDB will be exported to main.go
func ConnectDB() *sql.DB {
	// Connect to the Postgres Database
	db, err := sql.Open("postgres", "user=rodrigovalente password=password host=localhost port=5432 dbname=api_jwt sslmode=disable")

	if err != nil {
		panic(err)
	}

	return db
}
