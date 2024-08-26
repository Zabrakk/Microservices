package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetBasicAuthAllCorrect(t *testing.T) {
	r, _ := http.NewRequest("POST", "", bytes.NewReader([]byte("")))
	r.SetBasicAuth("test_user", "test_password")

	username, password, ok := GetBasicAuth(r)
	if !ok {
		t.Fatal("GetBasicAuth failed")
	}
	if username != "test_user" {
		t.Fatal("Username was incorrect", username)
	}
	if password != "test_password" {
		t.Fatal("Password was incorrect", password)
	}
}

func TestGetBasicAuthBasicAuthMissing(t *testing.T) {
	r, _ := http.NewRequest("POST", "", bytes.NewReader([]byte("")))

	_, _, ok := GetBasicAuth(r)
	if ok {
		t.Fatal("GetBasicAuth was ok, but BasicAuth header was missing")
	}
}

func TestGetBasicAuthCredentialsMissing(t *testing.T) {
	r, _ := http.NewRequest("POST", "", bytes.NewReader([]byte("")))
	r.SetBasicAuth("test_user", "")

	username, password, ok := GetBasicAuth(r)
	if ok { t.Fatal("GetBasicAuth failed") }
	if username != "" || password != "" {
		t.Fatal("Wrong result since password was empty", password)
	}

	r.SetBasicAuth("", "")

	username, password, ok = GetBasicAuth(r)
	if ok { t.Fatal("GetBasicAuth failed") }
	if username != "" || password != "" {
		t.Fatal("Wrong result since username and password were empty", password)
	}
}

func TestCreateJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret")
	tokenString, err := CreateJWT("test_user")
	if err != nil {
		t.Fatalf("An unexpected error occured while creating JWT:\n%s", err.Error())
	}
	if tokenString == "" {
		t.Fatal("tokenString was empty")
	}
}

func TestCreateJWTSecretNotSet(t *testing.T) {
	os.Setenv("JWT_SECRET", "")
	_, err := CreateJWT("test_user")
	if err == nil {
		t.Fatal("JWT was created with empty JWT_SECRET")
	}
}

func TestLogin(t *testing.T) {
	var mock sqlmock.Sqlmock
	var err error
	tests := []struct {
		name			string
		method			string
		expectedCode	int
		credentials		[]string
		row				[]string
		jwtSecret		string
	}{
		{
			name: "Successful login",
			method: "POST",
			expectedCode: 200,
			credentials: []string{"test_user", "test_password"},
			row: []string{"test_user", "test_password"},
			jwtSecret: "test_secret",
		},
		{
			name: "Incorrect HTTP request method",
			method: "GET",
			expectedCode: 405,
			credentials: []string{"test_user", "test_password"},
			row: []string{"test_user", "test_password"},
			jwtSecret: "test_secret",
		},
		{
			name: "Credentials missing",
			method: "POST",
			expectedCode: 401,
			credentials: []string{"test_user", ""},
			row: []string{"test_user", "test_password"},
			jwtSecret: "test_secret",
		},
		{
			name: "Username is incorrect",
			method: "POST",
			expectedCode: 401,
			credentials: []string{"test_user", "test_password"},
			row: []string{},
			jwtSecret: "test_secret",
		},
		{
			name: "Password is incorrect",
			method: "POST",
			expectedCode: 401,
			credentials: []string{"test_user", "test_password"},
			row: []string{"test_user", "different_password"},
			jwtSecret: "test_secret",
		},
		{
			name: "JWT creation fails",
			method: "POST",
			expectedCode: 500,
			credentials: []string{"test_user", "test_password"},
			row: []string{"test_user", "test_password"},
			jwtSecret: "",
		},
		{
			name: "DB fetch fails",
			method: "POST",
			expectedCode: 500,
			credentials: []string{"test_user", "test_password"},
			row: []string{"test_user", "test_password"},
			jwtSecret: "test_secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("JWT_SECRET", tt.jwtSecret)
			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			if tt.name == "DB fetch fails" {
				mock.ExpectQuery("SELECT email, password FROM user WHERE email=?").WithArgs(tt.row[0]).WillReturnError(errors.New("db fetch failed"))
			} else if len(tt.row) == 0 {
				mock.ExpectQuery("SELECT email, password FROM user WHERE email=?").WithArgs(tt.credentials[0]).WillReturnRows(sqlmock.NewRows([]string{"email", "password"}))
			} else {
				rows := sqlmock.NewRows([]string{"email", "password"}).AddRow(tt.row[0], tt.row[1])
				mock.ExpectQuery("SELECT email, password FROM user WHERE email=?").WithArgs(tt.row[0]).WillReturnRows(rows)
			}

			req, err := http.NewRequest(tt.method, "/login", nil)
			if err != nil { t.Fatalf("NewRequest creation failed:\n%s", err.Error()) }
			req.SetBasicAuth(tt.credentials[0], tt.credentials[1])

			resp := httptest.NewRecorder()
			handler := http.HandlerFunc(Login)
			handler.ServeHTTP(resp, req)

			if resp.Code != tt.expectedCode {
				t.Fatal("Status was incorrect", resp.Code)
			}
			if resp.Code == 200 {
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil { t.Fatalf("Error while reading resp body:\n%s", err.Error()) }
				if len(string(bodyBytes)) == 0 {
					t.Fatal("Did not receive JWT")
				}
			}
		})
	}
}

func TestRegister(t *testing.T) {
	var mock sqlmock.Sqlmock
	var err error
	tests := []struct {
		name			string
		method			string
		expectedCode	int
		credentials		[]string
		jwtSecret		string
	}{
		{
			name: "Successful registration",
			method: "POST",
			expectedCode: 200,
			credentials: []string{"test_user", "test_password"},
			jwtSecret: "test_secret",
		},
		{
			name: "Incorrect HTTP request method",
			method: "GET",
			expectedCode: 405,
			credentials: []string{"test_user", "test_password"},
			jwtSecret: "test_secret",
		},
		{
			name: "Credentials missing from headers",
			method: "POST",
			expectedCode: 400,
			credentials: []string{},
			jwtSecret: "test_secret",
		},
		{
			name: "JWT creation fails",
			method: "POST",
			expectedCode: 500,
			credentials: []string{"test_user", "test_password"},
			jwtSecret: "",
		},
		{
			name: "Duplicate in DB",
			method: "POST",
			expectedCode: 409,
			credentials: []string{"test_user", "test_password"},
			jwtSecret: "test_secret",
		},
		{
			name: "Insert into DB fails",
			method: "POST",
			expectedCode: 500,
			credentials: []string{"test_user", "test_password"},
			jwtSecret: "test_secret",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("JWT_SECRET", tt.jwtSecret)
			db, mock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			req, err := http.NewRequest(tt.method, "/register", nil)
			if err != nil { t.Fatalf("NewRequest creation failed:\n%s", err.Error()) }
			if len(tt.credentials) > 0 {
				req.Header.Add("Username", tt.credentials[0])
				req.Header.Add("Password", tt.credentials[1])
			}

			if tt.name == "Insert into DB fails" {
				mock.ExpectExec("INSERT INTO user (email, password) VALUES (?, ?)").WillReturnError(errors.New("Something bad happened"))
			} else if tt.name == "Duplicate in DB" {
				mock.ExpectExec("INSERT INTO user (email, password) VALUES (?, ?)").WillReturnError(errors.New("Error 1062 (23000): Duplicate entry"))
			} else if len(tt.credentials) > 0 {
				mock.ExpectExec("INSERT INTO user (email, password) VALUES (?, ?)").WithArgs(tt.credentials[0], tt.credentials[1]).WillReturnResult(sqlmock.NewResult(1, 1))
			}

			resp := httptest.NewRecorder()
			handler := http.HandlerFunc(Register)
			handler.ServeHTTP(resp, req)

			if status := resp.Code; status != tt.expectedCode {
				t.Fatal("Status was incorrect", status)
			}
			if resp.Code == 200 {
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil { t.Fatalf("Error while reading resp body:\n%s", err.Error()) }
				if len(string(bodyBytes)) == 0 {
					t.Fatal("Did not receive JWT")
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name			string
		method			string
		expectedCode	int
		jwtSecret		string
	} {
		{
			name: "Successful validation",
			method: "POST",
			expectedCode: 200,
			jwtSecret: "test_secret",
		},
		{
			name: "Incorrect HTTP request method",
			method: "GET",
			expectedCode: 405,
			jwtSecret: "test_secret",
		},
		{
			name: "JWT secret not set",
			method: "POST",
			expectedCode: 500,
			jwtSecret: "",
		},
		{
			name: "JWT missing from headers",
			method: "POST",
			expectedCode: 400,
			jwtSecret: "test_secret",
		},
		{
			name: "Bearer token length is incorrect",
			method: "POST",
			expectedCode: 400,
			jwtSecret: "test_secret",
		},
		{
			name: "Not Authorized",
			method: "POST",
			expectedCode: 403,
			jwtSecret: "test_secret",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("JWT_SECRET", tt.jwtSecret)
			req, err := http.NewRequest(tt.method, "/validate", nil)
			if err != nil { t.Fatalf("NewRequest creation failed:\n%s", err.Error()) }

			if tt.expectedCode == 200 {
				tokenString, err := CreateJWT("test_user")
				if err != nil { t.Fatalf("JWT creation failed\n:%s", err.Error()) }
				req.Header.Add("Authorization", "Bearer " +  tokenString)
			} else if tt.name == "Bearer token length is incorrect" {
				req.Header.Add("Authorization", "tokenString")
			} else if tt.expectedCode != 400 {
				req.Header.Add("Authorization", "Bearer " + "tokenString")
			}

			resp := httptest.NewRecorder()
			handler := http.HandlerFunc(Validate)
			handler.ServeHTTP(resp, req)

			if status := resp.Code; status != tt.expectedCode {
				t.Fatal("Status was incorrect", status)
			}
			if resp.Code == 200 {
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil { t.Fatalf("Error while reading resp body:\n%s", err.Error()) }
				body := string(bodyBytes)
				if len(body) == 0 {
					t.Fatal("Did not receive JWT")
				}
				if !strings.Contains(body, `"username":"test_user"`) {
					t.Fatal("Response JSON did not include username")
				}
				if !strings.Contains(body, `"admin":true}`) {
					t.Fatal("Response JSON did not include admin")
				}
			}
		})
	}
}