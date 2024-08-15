package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

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
func CreateJWT(username string) (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp": time.Now().Add(time.Hour * 24).Unix(),
		})
	tokenString, err = token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// This method is used to send a HTTP response with status code 405.
// Use it when receiving a request with an unallowed HTTP method.
func SendMethodNotAllowed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintf(w, "The method is not allowed for the requested URL.")
}

func SendInvalidCredentials(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, "Credentials were invalid")
}

func GetBasicAuth(w http.ResponseWriter, r *http.Request) (username string, password string, ok bool) {
	username, password, ok = r.BasicAuth(); if !ok {
		log.Println("BasicAuth() was not ok")
		SendInvalidCredentials(w)
		return "", "", ok
	}
	if len(username) == 0 || len(password) == 0 {
		log.Println("Username or password was empty")
		SendInvalidCredentials(w)
		return "", "", false
	}
	return username, password, ok
}

// TODO
func Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Login request received with method", r.Method)
	if r.Method != "POST" {
		SendMethodNotAllowed(w)
		return
	}
	username, password, ok := GetBasicAuth(w, r); if !ok { return }
	log.Printf("Log in request received for user %s with password %s\n", username, password)
	rows, err := db.Query(`SELECT email, password FROM user WHERE email=?`, username)
	if err != nil {
		log.Println("Error occured while trying to fetch user from DB")
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal server error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r_user string
		var r_password string
		if err := rows.Scan(&r_user, &r_password); err != nil {
			log.Println("Error occured while trying to fetch user from DB")
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Internal server error")
			return
		}
		log.Printf("Read %s %s from db\n", r_user, r_password)

		if username != r_user || password != r_password {
			SendInvalidCredentials(w)
			return
		} else {
			tokenString, err := CreateJWT(r_user)
			if err != nil {
				log.Println("Error occured while trying tocreate JWT")
				log.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Internal server error")
				return
			}
			fmt.Fprintf(w, "%s", tokenString)
			return
		}
	}
	SendInvalidCredentials(w)
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
	var err error

	// Connect to MySQL
	mySqlConf := NewMySQLConf()
	db, err = sql.Open("mysql", mySqlConf.GetDataSourceName())
	if err != nil {
		log.Panic(err.Error())
	}
	defer db.Close()

	// Register handler functions to routes
	http.HandleFunc("/login", Login)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/validate", Validate)

	servicePort := os.Getenv("SERVICE_PORT")

	log.Println("Authorization service running on port", servicePort)

	err = http.ListenAndServe(":"+servicePort, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}