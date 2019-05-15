package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/user/REST_API_JWT/controllers"
	"github.com/user/REST_API_JWT/driver"
	"github.com/user/REST_API_JWT/middlewares"
	"github.com/user/REST_API_JWT/utils"
)

var db *sql.DB

func main() {

	db = driver.ConnectDB()

	// gorilla/mux
	router := mux.NewRouter()
	router.HandleFunc("/signup", utils.Logging(controllers.Signup)).Methods("POST")
	router.HandleFunc("/login", utils.Logging(controllers.Login)).Methods("POST")
	router.HandleFunc("/protected", middlewares.TokenVerifyMiddleware(utils.Logging(controllers.ProtectedEndpoint))).Methods("GET")

	log.Println("Listen on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", router))
}
