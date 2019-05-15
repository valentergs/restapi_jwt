package controllers

import (
	"fmt"
	"net/http"
)

//ProtectedEndpoint will be exported
func ProtectedEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println("protectedEndpoint invoked.")
}
