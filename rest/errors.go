package rest

import "net/http"

// RestError is an implementation of the error interface representing an HTTP error.
type RestError struct {
	reason string
	status int
}

// Error returns the RestError message.
func (r RestError) Error() string { return r.reason }

// Status returns the HTTP status code.
func (r RestError) Status() int { return r.status }

// ResourceNotFound returns a RestError for a 404 Not Found error.
func ResourceNotFound(message ...string) RestError {
	reason := "Resource not found"
	if len(message) > 0 {
		reason = message[0]
	}
	return RestError{reason, http.StatusNotFound}
}

// ResourceNotPermitted returns a RestError for a 403 Forbidden error.
func ResourceNotPermitted(message ...string) RestError {
	reason := "Resource forbidden"
	if len(message) > 0 {
		reason = message[0]
	}
	return RestError{reason, http.StatusForbidden}
}

// ResourceConflict returns a RestError for a 409 Conflict error.
func ResourceConflict(message ...string) RestError {
	reason := "Resource conflict"
	if len(message) > 0 {
		reason = message[0]
	}
	return RestError{reason, http.StatusConflict}
}

// BadRequest returns a RestError for a 400 Bad Request error.
func BadRequest(message ...string) RestError {
	reason := "Bad request"
	if len(message) > 0 {
		reason = message[0]
	}
	return RestError{reason, http.StatusBadRequest}
}

// UnprocessableRequest returns a RestError for a 422 Unprocessable Entity error.
func UnprocessableRequest(message ...string) RestError {
	reason := "Unprocessable request"
	if len(message) > 0 {
		reason = message[0]
	}
	return RestError{reason, 422}
}

// UnauthorizedRequest returns a RestError for a 401 Unauthorized error.
func UnauthorizedRequest(message ...string) RestError {
	reason := "Unauthorized request"
	if len(message) > 0 {
		reason = message[0]
	}
	return RestError{reason, http.StatusUnauthorized}
}

// NotImplemented returns a RestError for a 501 Not Implemented error.
func NotImplemented(message ...string) RestError {
	reason := "Not implemented"
	if len(message) > 0 {
		reason = message[0]
	}
	return RestError{reason, http.StatusNotImplemented}
}

// InternalServerError returns a RestError for a 500 Internal Server error.
func InternalServerError(message ...string) RestError {
	reason := "Internal server error"
	if len(message) > 0 {
		reason = message[0]
	}
	return RestError{reason, http.StatusInternalServerError}
}
