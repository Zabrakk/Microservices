package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func MockAuthServiceHandler(w http.ResponseWriter, r *http.Request) {
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
				mockAuthService := httptest.NewServer(http.HandlerFunc(MockAuthServiceHandler))
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

func TestRegister(t *testing.T) {
	tests := []struct {
		name			string
		method			string
		expectedCode	int
		credentials		[]string
	} {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/register", nil)
			if err != nil { t.Fatalf("NewRequest creation failed:\n%s", err.Error()) }
			req.Header.Add("Username", tt.credentials[0])
			req.Header.Add("Password", tt.credentials[1])

			resp := httptest.NewRecorder()
			handler := http.HandlerFunc(Register)
			handler.ServeHTTP(resp, req)

			if resp.Code != tt.expectedCode { t.Fatal("Status was incorrect", resp.Code) }
		})
	}
}