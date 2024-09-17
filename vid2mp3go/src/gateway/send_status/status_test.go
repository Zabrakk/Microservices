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

func TestBasedOnValue(t *testing.T) {
	w := MockResponseWriter{}
	tests := []struct{
		statusCode	int
	}{
		{ statusCode: 200 },
		{ statusCode: 400 },
		{ statusCode: 401 },
		{ statusCode: 403 },
		{ statusCode: 405 },
		{ statusCode: 409 },
		{ statusCode: 500 },
	}
	for _, tt := range tests {
		ok := BasedOnValue(&w, tt.statusCode)
		if tt.statusCode != 200 {
			if w.Code != tt.statusCode {
				t.Fatal("Status code was incorrect", w.Code)
			}
			if ok {
				t.Fatal("ok should have been false")
			}
		} else {
			if w.Code != 0 {
				t.Fatal("Status code was incorrect", w.Code)
			}
			if !ok {
				t.Fatal("ok should have been true")
			}
		}

	}
}