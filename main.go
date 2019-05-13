package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	//"github.com/gorilla/mux"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

//User is an exportable struct
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

//JWT is an exportable struct
type JWT struct {
	Token string `json:"token"`
}

//Error is an exportable struct
type Error struct {
	Message string `json:"message"`
}

var db *sql.DB

func main() {

	//declare err here since we cannot declare it below with the short-hand declaration syntax
	//that will give us an error because db is already decalred outside the var keyword, and it does not need to be
	//to be re-declared.
	var err error

	// Connect to the Postgres Database
	db, err = sql.Open("postgres", "user=rodrigovalente password=password host=localhost port=5432 dbname=api_jwt sslmode=disable")

	if err != nil {
		panic(err)
	}
	// Close the connection
	defer db.Close()

	// gorilla/mux
	router := mux.NewRouter()
	router.HandleFunc("/signup", signup).Methods("POST")
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/protected", TokenVerifyMiddleware(protectedEndpoint)).Methods("GET")

	// // // julienschmidt/httprouter
	// router := httprouter.New()
	// router.POST("/signup", signup)
	// router.POST("/login", login)
	// //router.GET("/protected", TokenVerifyMiddleware(protectedEndpoint))

	log.Println("Listen on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", router))

}

func respondWithError(w http.ResponseWriter, status int, error Error) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(error)
	return
}

func responseJSON(w http.ResponseWriter, data interface{}) {
	json.NewEncoder(w).Encode(data)
}

func signup(w http.ResponseWriter, r *http.Request) {
	var user User
	var error Error

	json.NewDecoder(r.Body).Decode(&user)

	if user.Email == "" {
		error.Message = "Favor inserir um e-mail válido"
		respondWithError(w, http.StatusBadRequest, error)
		return
	}
	if user.Password == "" {
		error.Message = "Favor inserir uma senha válida"
		respondWithError(w, http.StatusBadRequest, error)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)

	if err != nil {
		log.Fatal(err)
	}

	user.Password = string(hash)

	stmt := "insert into users (email, password) values($1, $2) RETURNING id;"

	err = db.QueryRow(stmt, user.Email, user.Password).Scan(&user.ID)

	if err != nil {
		error.Message = "Server error."
		respondWithError(w, http.StatusInternalServerError, error)
		return
	}

	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	responseJSON(w, user)
}

func login(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("login invoked.")
	w.Write([]byte("Acesso ao login!"))
}

func protectedEndpoint(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("protectedEndpoint invoked.")
}

func TokenVerifyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	fmt.Println("TokenVerifyMiddleware invoked.")
	return nil
}
