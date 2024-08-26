package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	MySQLConf "microservices/authorization/mysql_conf"
	SendStatus "microservices/authorization/send_status"
	"net/http"
	"os"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type JsonStruct struct {
	Username	string		`json:"username"`
	Exp			float64		`json:"exp"`
	Admin		bool		`json:"admin"`
}

// Gets the BasicAuth credentials present in a given http.Request.
// If there are no credentials present, "ok" will be false.
// If everything is ok, the username and password are returned.
func GetBasicAuth(r *http.Request) (username string, password string, ok bool) {
	username, password, ok = r.BasicAuth(); if !ok {
		return "", "", false
	}
	if len(username) == 0 || len(password) == 0 {
		return "", "", false
	}
	return username, password, ok
}

// Returns JWT string, expiring in one day, for a given user.
// If something goes wrong, an error is returned.
func CreateJWT(username string) (tokenString string, err error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("env variable JWT_SECRET was empty")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp": time.Now().Add(time.Hour * 24).Unix(),
			"admin": true,
		})
	tokenString, err = token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Login handler. Checks that credentials present in the http.Request's
// BasicAuth header also exist in the Authorization database.
// A JWT is returned on successful login, otherwise an error is returned.
func Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Login request received with method", r.Method)
	if r.Method != "POST" {
		SendStatus.MethodNotAllowed(w)
		return
	}
	username, password, ok := GetBasicAuth(r); if !ok {
		SendStatus.InvalidCredentials(w)
		return
	}
	log.Printf("Log in request received for user %s with password %s\n", username, password)
	rows, err := db.Query(`SELECT email, password FROM user WHERE email=?`, username)
	if err != nil {
		log.Printf("Error occured while trying to fetch user from DB:\n%s", err.Error())
		SendStatus.InternalServerError(w)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r_user, r_password string
		if err := rows.Scan(&r_user, &r_password); err != nil {
			log.Printf("Error occured while trying to fetch user from DB:\n%s", err.Error())
			SendStatus.InternalServerError(w)
			return
		}
		log.Printf("Read %s %s from db\n", r_user, r_password)

		if username != r_user || password != r_password {
			SendStatus.InvalidCredentials(w)
			return
		} else {
			tokenString, err := CreateJWT(r_user)
			if err != nil {
				log.Printf("Error occured while trying to create JWT:\n%s", err.Error())
				SendStatus.InternalServerError(w)
				return
			}
			fmt.Fprintf(w, "%s", tokenString)
			return
		}
	}
	SendStatus.InvalidCredentials(w)
}

// Attempts to register a new user based on the Username and Password included in the
// received POST request's headers. A JWT is returned after successful registrations.
// In all other cases, an error is returned.
func Register(w http.ResponseWriter, r *http.Request) {
	log.Println("Register request received with method", r.Method)
	if r.Method != "POST" {
		SendStatus.MethodNotAllowed(w)
		return
	}
	username, password := r.Header.Get("Username"), r.Header.Get("Password")
	if username == "" || password == "" {
		SendStatus.BadRequest(w)
		return
	}
	_, err := db.Exec("INSERT INTO user (email, password) VALUES (?, ?)", username, password)

	if err != nil {
		errString := err.Error()
		log.Printf("Something went wrong trying to register user to DB:\n%s", errString)
		if strings.Contains(errString, "Error 1062") {
			SendStatus.Conflict(w)
			return
		}
		SendStatus.InternalServerError(w)
		return
	}
	tokenString, err := CreateJWT(username)
	if err != nil {
		log.Printf("Error occured while trying to create JWT:\n%s", err.Error())
		SendStatus.InternalServerError(w)
		return
	}
	fmt.Fprintf(w, "%s", tokenString)
}

// Checks whether a valid JSON Web Token is present in the received POST request.
func Validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		SendStatus.MethodNotAllowed(w)
		return
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("Env variable JWT_SECRET was empty")
		SendStatus.InternalServerError(w)
		return
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("JWT was missing from request headers")
		SendStatus.BadRequest(w)
		return
	}
	tokenString := strings.Split(authHeader, " ")
	if len(tokenString) != 2 {
		log.Println("JWT length was incorrect after split")
		SendStatus.BadRequest(w)
		return
	}
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString[1], claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		log.Printf("JWT decode failed:\n%s", err.Error())
		SendStatus.Forbidden(w)
		return
	}
	res := JsonStruct{}
	for key, val := range claims {
		if key == "username" {
			res.Username = val.(string)
		} else if key == "admin" {
			res.Admin = val.(bool)
		} else if key == "exp" {
			res.Exp = val.(float64)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func main() {
	log.Println("Authorization service starting...")
	var err error

	// Connect to MySQL
	mySqlConf := MySQLConf.NewMySQLConf()
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