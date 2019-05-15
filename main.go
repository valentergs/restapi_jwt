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
	"github.com/user/rest_api_jwt/driver"
	"github.com/user/rest_api_jwt/models"
	"github.com/user/rest_api_jwt/utils"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func main() {

	db = driver.ConnectDB()

	// gorilla/mux
	router := mux.NewRouter()
	router.HandleFunc("/signup", utils.Logging(signup)).Methods("POST")
	router.HandleFunc("/login", utils.Logging(login)).Methods("POST")
	router.HandleFunc("/protected", TokenVerifyMiddleware(utils.Logging(protectedEndpoint))).Methods("GET")

	log.Println("Listen on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", router))

}

func signup(w http.ResponseWriter, r *http.Request) {
	var user models.User
	var error models.Error

	json.NewDecoder(r.Body).Decode(&user)

	if user.Email == "" {
		error.Message = "Favor inserir um e-mail válido"
		utils.RespondWithError(w, http.StatusBadRequest, error)
		return
	}
	if user.Password == "" {
		error.Message = "Favor inserir uma senha válida"
		utils.RespondWithError(w, http.StatusBadRequest, error)
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
		utils.RespondWithError(w, http.StatusInternalServerError, error)
		return
	}

	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	utils.ResponseJSON(w, user)
}

//GenerateToken is an exportable function
func GenerateToken(user models.User) (string, error) {
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
	var user models.User
	var jwt models.JWT
	var error models.Error

	json.NewDecoder(r.Body).Decode(&user)

	if user.Email == "" {
		error.Message = "Email missing"
		utils.RespondWithError(w, http.StatusBadRequest, error)
		return
	}

	if user.Password == "" {
		error.Message = "Password missing"
		utils.RespondWithError(w, http.StatusBadRequest, error)
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
			utils.RespondWithError(w, http.StatusBadRequest, error)
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
		utils.RespondWithError(w, http.StatusUnauthorized, error)
		return
	}

	w.WriteHeader(http.StatusOK)

	jwt.Token = token

	utils.ResponseJSON(w, jwt)
}

func protectedEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println("protectedEndpoint invoked.")
}

//TokenVerifyMiddleware will validate the token that was sent by the user giving access to the "protectedend point".  It takes "next" as an argument - it is triggered after the token is validated.
func TokenVerifyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

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
				utils.RespondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}

			// if the token is valid, next will call the next function which is "protectedEndpoint"
			if token.Valid {
				next.ServeHTTP(w, r)
			} else {
				errorObject.Message = error.Error()
				utils.RespondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}
		} else {
			errorObject.Message = "Invalid token."
			utils.RespondWithError(w, http.StatusUnauthorized, errorObject)
			return
		}
	})
}
