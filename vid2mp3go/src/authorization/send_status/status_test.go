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

type fn func(http.ResponseWriter)

func CheckStatus(f fn, expectedStatus int, t *testing.T) {
	w := MockResponseWriter{}
	f(&w)
	if w.Code != expectedStatus {
		t.Fatal("Status code was incorrect", w.Code)
	}
}

func TestBadRequest(t *testing.T) {
	CheckStatus(BadRequest, 400, t)
}

func TestInvalidCredentials(t *testing.T) {
	CheckStatus(InvalidCredentials, 401, t)
}

func TestForbidden(t *testing.T) {
	CheckStatus(Forbidden, 403, t)
}

func TestMethodNotwAllowed(t *testing.T) {
	CheckStatus(MethodNotAllowed, 405, t)
}

func TestConflict(t *testing.T) {
	CheckStatus(Conflict, 409, t)
}

func TestInternalServerError(t *testing.T) {
	CheckStatus(InternalServerError, 500, t)
}