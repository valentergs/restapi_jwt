package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
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
	log.Printf("%s %s %v", r.URL, r.Method, r.Proto)
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

	// this is the user.Password variable result of the DB query
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
	log.Printf("%s %s %v", r.URL, r.Method, r.Proto)
}

func protectedEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println("protectedEndpoint invoked.")
	log.Printf("%s %s %v", r.URL, r.Method, r.Proto)
}

//TokenVerifyMiddleware will validate the token that was sent by the user giving access to the "protectedend point".  It takes "next" as an argument - it is triggered after the token is validated.
func TokenVerifyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errorObject Error

		// this header should have a key/value pair called "Authorization". "authHeader" will grab the key
		authHeader := r.Header.Get("Authorization")
		// bearerToken will remove the empty space found on the value
		bearerToken := strings.Split(authHeader, " ")

		if len(bearerToken) == 2 {
			// here we catch the value of bearerToken[1] leaving the word "bearer" out.
			authToken := bearerToken[1]

			// to make sure the token is valid we use jwt.Parse
			token, error := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return []byte("secret"), nil
			})

			if error != nil {
				errorObject.Message = error.Error()
				respondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}

			// if the token is valid, next will call the next function which is "protectedEndpoint"
			if token.Valid {
				next.ServeHTTP(w, r)
			} else {
				errorObject.Message = error.Error()
				respondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}
		} else {
			errorObject.Message = "Invalid token."
			respondWithError(w, http.StatusUnauthorized, errorObject)
			return
		}
	})
}
