package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/user/REST_API_JWT/models"
)

//RespondWithError will be exported
func RespondWithError(w http.ResponseWriter, status int, error models.Error) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(error)
	return
}

//ResponseJSON will be exported
func ResponseJSON(w http.ResponseWriter, data interface{}) {
	json.NewEncoder(w).Encode(data)
}

//Logging Com essa função emitimos um log para o terminal informanto status do servidor. Para cada handler do mux precisa passar essa função que tem como argumento as demais funções de handler.
func Logging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %v", r.URL, r.Method, r.Proto)
		f(w, r)
	}
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
