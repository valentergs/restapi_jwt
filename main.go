package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"

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

//GenerateToken is an exportable function
func GenerateToken(user User) (string, error) {
	var err error
	secret := "secret"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"iss":   "course",
	})

	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		log.Fatal(err)
	}

	return tokenString, nil
}

func login(w http.ResponseWriter, r *http.Request) {
	var user User
	var jwt JWT
	var error Error

	json.NewDecoder(r.Body).Decode(&user)

	if user.Email == "" {
		error.Message = "Email missing"
		respondWithError(w, http.StatusBadRequest, error)
		return
	}

	if user.Password == "" {
		error.Message = "Password missing"
		respondWithError(w, http.StatusBadRequest, error)
		return
	}

	//This varialbe will provided by the user when login
	password := user.Password

	// Check the database if the provided user exists in the table
	row := db.QueryRow("select * from users where email=$1;", user.Email)
	err := row.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			error.Message = "The user does not exists"
			respondWithError(w, http.StatusBadRequest, error)
			return
		} else {
			log.Fatal(err)
		}
	}

	token, err := GenerateToken(user)
	if err != nil {
		log.Fatal(err)
	}

	hashedPassword := user.Password

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		error.Message = "Invalid Password"
		respondWithError(w, http.StatusUnauthorized, error)
		return
	}

	w.WriteHeader(http.StatusOK)

	jwt.Token = token

	responseJSON(w, jwt)
}

func protectedEndpoint(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("protectedEndpoint invoked.")
}

func TokenVerifyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	fmt.Println("TokenVerifyMiddleware invoked.")
	return nil
}
