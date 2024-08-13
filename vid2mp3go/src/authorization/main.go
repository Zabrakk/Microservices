package main

import (
	"fmt"
	"log"
	"net/http"
)

// TODO: Get port from environment variable
const portNum string = "8080"

func CreateJWT() {

}

func Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Login request received with method", r.Method)
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "The method is not allowed for the requested URL.")
	}
	fmt.Fprintf(w, "Logged in")
}

func Register(w http.ResponseWriter, r *http.Request) {

}

func Validate(w http.ResponseWriter, r *http.Request) {

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