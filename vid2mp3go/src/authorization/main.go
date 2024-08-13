package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// TODO: Get port from environment variable
const portNum string = "8080"

type MySQLConf struct {
	Host		string
	DB			string
	Port		string
	User		string
	Password	string
}

func NewMySQLConf() MySQLConf {
	return MySQLConf{
		Host: os.Getenv("MYSQL_HOST"),
		DB: os.Getenv("MYSQL_DB"),
		Port: os.Getenv("MYSQL_PORT"),
		User: os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
	}
}

func (c MySQLConf) GetDataSourceName() string {
	return  c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + c.Port + ")/" + c.DB
}

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

	// Connect to MySQL
	mySqlConf := NewMySQLConf()
	db, err := sql.Open("mysql", mySqlConf.GetDataSourceName())
	if err != nil {
		log.Panic(err.Error())
	}
	defer db.Close()

	// Register handler functions to routes
	http.HandleFunc("/login", Login)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/validate", Validate)

	log.Println("Authorization service running on port", portNum)

	err = http.ListenAndServe(":"+portNum, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}