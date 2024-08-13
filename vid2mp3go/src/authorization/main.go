package main

import (
	"fmt"
	"log"
	"net/http"
)

// TODO: Get port from environment variable
const portNum string = "8080"

// Used to create JWT tokens for users.
func CreateJWT() {

}

// This method is used to send a HTTP response with status code 405.
// Use it when receiving a request with an unallowed HTTP method.
func SendMethodNotAllowed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintf(w, "The method is not allowed for the requested URL.")
}

// TODO
func Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Login request received with method", r.Method)
	if r.Method != "POST" {
		SendMethodNotAllowed(w)
	} else {
		fmt.Fprintf(w, "Logged in")
	}
}

// TODO
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		SendMethodNotAllowed(w)
	} else {
		fmt.Fprintf(w, "Registered")
	}
}

// TODO
func Validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		SendMethodNotAllowed(w)
	} else {
		fmt.Fprintf(w, "Validated")
	}
}

func main() {
	log.Println("Authorization service starting...")

	// Register handler functions to routes
	http.HandleFunc("/login", Login)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/validate", Validate)

	log.Println("Authorization service running on port", portNum)

	err := http.ListenAndServe(":"+portNum, nil)
	if err != nil {
		log.Fatal(err)
	}
}