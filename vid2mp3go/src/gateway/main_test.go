package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func MockLoginHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || username != "test" || password != "test" {
		w.WriteHeader(401)
	}
	w.Write([]byte("tokenString"))
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name			string
		method			string
		expectedCode	int
		credentials		[]string
	}{
		{
			name: "Successful login",
			method: "POST",
			expectedCode: 200,
			credentials: []string{"test", "test"},
		},
		{
			name: "Incorrect HTTP request method",
			method: "GET",
			expectedCode: 405,
			credentials: []string{"test", "test"},
		},
		{
			name: "Credentials missing",
			method: "POST",
			expectedCode: 401,
			credentials: []string{"", ""},
		},
		{
			name: "Auth service not reachable",
			method: "POST",
			expectedCode: 500,
			credentials: []string{"test", "test"},
		},
		{
			name: "Credentials incorrect",
			method: "POST",
			expectedCode: 401,
			credentials: []string{"wrong", "wrong"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/login", nil)
			if err != nil { t.Fatalf("NewRequest creation failed:\n%s", err.Error()) }
			req.SetBasicAuth(tt.credentials[0], tt.credentials[1])

			if tt.expectedCode != 500 {
				// When expectedCode is 500 the AuthService should not be reachable.
				mockAuthService := httptest.NewServer(http.HandlerFunc(MockLoginHandler))
				defer mockAuthService.Close()
				GetAuthServiceUrl = func() (url string) { return mockAuthService.URL }
			}

			resp := httptest.NewRecorder()
			handler := http.HandlerFunc(Login)
			handler.ServeHTTP(resp, req)

			if resp.Code != tt.expectedCode { t.Fatal("Status was incorrect", resp.Code) }
			if resp.Code == 200 {
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil { t.Fatalf("Error while reading resp body:\n%s", err.Error()) }
				if string(bodyBytes) != "tokenString" {
					t.Fatal("Did not receive JWT")
				}
			}
		})
	}
}

func MockRegisterHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("Username")
	password := r.Header.Get("Password")
	if username == "" || password == "" {
		w.WriteHeader(400)
	}
	w.Write([]byte("tokenString"))
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name			string
		method			string
		expectedCode	int
		credentials		[]string
	} {
		{
			name: "Successful registration",
			method: "POST",
			expectedCode: 200,
			credentials: []string{"test", "test"},
		},
		{
			name: "Incorrect HTTP request method",
			method: "GET",
			expectedCode: 405,
			credentials: []string{"test", "test"},
		},
		{
			name: "Credentails are missing",
			method: "POST",
			expectedCode: 400,
			credentials: []string{"", ""},
		},
		{
			name: "Auth service not reachable",
			method: "POST",
			expectedCode: 500,
			credentials: []string{"test", "test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/register", nil)
			if err != nil { t.Fatalf("NewRequest creation failed:\n%s", err.Error()) }
			req.Header.Add("Username", tt.credentials[0])
			req.Header.Add("Password", tt.credentials[1])

			if tt.expectedCode != 500 {
				// When expectedCode is 500 the AuthService should not be reachable.
				mockAuthService := httptest.NewServer(http.HandlerFunc(MockRegisterHandler))
				defer mockAuthService.Close()
				GetAuthServiceUrl = func() (url string) { return mockAuthService.URL }
			}

			resp := httptest.NewRecorder()
			handler := http.HandlerFunc(Register)
			handler.ServeHTTP(resp, req)

			if resp.Code != tt.expectedCode { t.Fatal("Status was incorrect", resp.Code) }
			if resp.Code == 200 {
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil { t.Fatalf("Error while reading resp body:\n%s", err.Error()) }
				if string(bodyBytes) != "tokenString" {
					t.Fatal("Did not receive JWT")
				}
			}
		})
	}
}

func MockValidationHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	tokenString := strings.Split(auth, " ")
	if len(tokenString) != 2 {
		w.WriteHeader(400)
		return
	}
	if tokenString[1] != "test" {
		w.WriteHeader(403)
		return
	}
	w.Write([]byte("jwt"))
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name			string
		expectedCode	int
		header			string
	}{
		{
			name: "Successful validation",
			expectedCode: 200,
			header: "bearer test",
		},
		{
			name: "Auth header empty or missing",
			expectedCode: 401,
			header: "",
		},
		{
			name: "Auth header incorrect format",
			expectedCode: 400,
			header: "test",
		},
		{
			name: "Auth service not reachable",
			expectedCode: 500,
			header: "test",
		},
	}
	for _, tt := range(tests) {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedCode != 500 {
				mockAuthService := httptest.NewServer(http.HandlerFunc(MockValidationHandler))
				defer mockAuthService.Close()
				GetAuthServiceUrl = func() (url string) { return mockAuthService.URL }
			}

			req, err := http.NewRequest("POST", "", nil)
			if err != nil { t.Fatalf("NewRequest creation failed:\n%s", err.Error()) }
			req.Header.Set("Authorization", tt.header)

			jwtObject, statusCode := ValidateToken(req)
			if statusCode != tt.expectedCode { t.Fatal("Status was incorrect", statusCode) }
			if statusCode == 200 {
				if string(jwtObject) != "jwt" { t.Fatal("Received JWT object was incorrect") }
			}


		})
	}
}
