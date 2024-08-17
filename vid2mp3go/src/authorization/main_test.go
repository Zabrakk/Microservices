package main

import (
	"bytes"
	"net/http"
	"testing"
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
