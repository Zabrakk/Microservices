package sendstatus

import (
	"net/http"
	"testing"
)

type MockResponseWriter struct {
	Code int
}

func (w *MockResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *MockResponseWriter) Write(data []byte) (int, error) {
	return 1, nil
}

func (w *MockResponseWriter) WriteHeader(statusCode int) {
	w.Code = statusCode
}

func TestBadRequest(t *testing.T) {
	w := MockResponseWriter{}
	BadRequest(&w)
	if w.Code != 400 {
		t.Fatal("Status code was incorrect", w.Code)
	}
}

func TestInvalidCredentials(t *testing.T) {
	w := MockResponseWriter{}
	InvalidCredentials(&w)
	if w.Code != 401 {
		t.Fatal("Status code was incorrect", w.Code)
	}
}

func TestMethodNowAllowed(t *testing.T) {
	w := MockResponseWriter{}
	MethodNotAllowed(&w)
	if w.Code != 405 {
		t.Fatal("Status code was incorrect", w.Code)
	}
}

func TestConflict(t *testing.T) {
	w := MockResponseWriter{}
	Conflict(&w)
	if w.Code != 409 {
		t.Fatal("Status code was incorrect", w.Code)
	}
}

func TestInternalServerError(t *testing.T) {
	w := MockResponseWriter{}
	InternalServerError(&w)
	if w.Code != 500 {
		t.Fatal("Status code was incorrect", w.Code)
	}
}