package main

import (
	"fmt"
	"os"
	"testing"
)

func TestGetMongUri(t *testing.T) {
	tests := []struct {
		name	string
		vals	[]string
	}{
		{ name: "Normal run", vals: []string{"host", "1234"}, },
		{ name: "Env missing", vals: []string{"", ""}, },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MONGODB_HOST", tt.vals[0])
			os.Setenv("MONGODB_PORT", tt.vals[1])
			if len(tt.vals[0]) != 0 {
				uri, err := GetMongoUri()
				if err != nil { t.Fatal(err.Error()) }
				if uri != fmt.Sprintf("mongodb://%s:%s", tt.vals[0], tt.vals[1]) {
					t.Fatal("MongoUri was incorrect:", uri)
				}
			} else {
				_, err := GetMongoUri()
				if err == nil { t.Fatal("No error returned when Env was empty") }
			}
		})
	}
}