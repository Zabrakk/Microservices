package main

import (
	"log"
	"net/http"
)

var servicePort string = "8080"

func Login(w http.ResponseWriter, r *http.Request) {

}

func Register(w http.ResponseWriter, r *http.Request) {

}

func Upload(w http.ResponseWriter, r *http.Request) {

}

func Download(w http.ResponseWriter, r *http.Request) {

}

func main() {
	log.Println("Gateway service starting...")

	http.HandleFunc("/login", Login)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/upload", Upload)
	http.HandleFunc("/download", Download)

	log.Println("Gateway service running on port", servicePort)
	err := http.ListenAndServe(":"+servicePort, nil)
	if err != nil { log.Fatal(err.Error()) }
}