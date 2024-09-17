package sendstatus

import (
	"fmt"
	"net/http"
)

// This function is used to send a HTTP response with status code 400.
// Use it, e.g. when required headers are missing from a request.
func BadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "Bad request.")
}

// This function is used to send a HTTP response with status code 401.
// Use it when receiving a request with invalid credentials.
func InvalidCredentials(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, "Credentials were invalid.")
}

// This function is used to send a HTTP response with status code 403.
// Use it, e.g., when someone is not authorized
func Forbidden(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprintf(w, "Credentials were invalid.")
}

// This function is used to send a HTTP response with status code 405.
// Use it when receiving a request with an unallowed HTTP method.
func MethodNotAllowed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintf(w, "The method is not allowed for the requested URL.")
}

// This function is used to send a HTTP response with status code 409.
// Use, e.g., when a duplicate entry error occurs with the DB.
func Conflict(w http.ResponseWriter) {
	w.WriteHeader(http.StatusConflict)
	fmt.Fprintf(w, "Conflict.")
}

// This function is used to send a HTTP response with status code 500.
// Use it when something unexpected occurs.
func InternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Internal server error.")
}

// Sends out the appropriate status and message based on given statusCode.
// Returns false if one of the known error responses was sent.
func BasedOnValue(w http.ResponseWriter, statusCode int) (ok bool) {
	switch statusCode {
	case 400:
		BadRequest(w)
		return false
	case 401:
		InvalidCredentials(w)
		return false
	case 403:
		Forbidden(w)
		return false
	case 405:
		MethodNotAllowed(w)
		return false
	case 409:
		Conflict(w)
		return false
	case 500:
		InternalServerError(w)
		return false
	default:
		return true
	}
}