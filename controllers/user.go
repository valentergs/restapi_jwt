package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/user/REST_API_JWT/models"
	"github.com/user/REST_API_JWT/utils"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

//Signup will be exported
func Signup(w http.ResponseWriter, r *http.Request) {
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

func Login(w http.ResponseWriter, r *http.Request) {
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

	token, err := utils.GenerateToken(user)
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
