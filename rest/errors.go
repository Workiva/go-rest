package rest

import "net/http"

// Error is an implementation of the error interface representing an HTTP error.
type Error struct {
	reason string
	status int
}

// Error returns the Error message.
func (r Error) Error() string { return r.reason }

// Status returns the HTTP status code.
func (r Error) Status() int { return r.status }

// ResourceNotFound returns a Error for a 404 Not Found error.
func ResourceNotFound(message ...string) Error {
	reason := "Resource not found"
	if len(message) > 0 {
		reason = message[0]
	}
	return Error{reason, http.StatusNotFound}
}

// ResourceNotPermitted returns a Error for a 403 Forbidden error.
func ResourceNotPermitted(message ...string) Error {
	reason := "Resource forbidden"
	if len(message) > 0 {
		reason = message[0]
	}
	return Error{reason, http.StatusForbidden}
}

// ResourceConflict returns a Error for a 409 Conflict error.
func ResourceConflict(message ...string) Error {
	reason := "Resource conflict"
	if len(message) > 0 {
		reason = message[0]
	}
	return Error{reason, http.StatusConflict}
}

// BadRequest returns a Error for a 400 Bad Request error.
func BadRequest(message ...string) Error {
	reason := "Bad request"
	if len(message) > 0 {
		reason = message[0]
	}
	return Error{reason, http.StatusBadRequest}
}

// UnprocessableRequest returns a Error for a 422 Unprocessable Entity error.
func UnprocessableRequest(message ...string) Error {
	reason := "Unprocessable request"
	if len(message) > 0 {
		reason = message[0]
	}
	return Error{reason, 422}
}

// UnauthorizedRequest returns a Error for a 401 Unauthorized error.
func UnauthorizedRequest(message ...string) Error {
	reason := "Unauthorized request"
	if len(message) > 0 {
		reason = message[0]
	}
	return Error{reason, http.StatusUnauthorized}
}

// NotImplemented returns a Error for a 501 Not Implemented error.
func NotImplemented(message ...string) Error {
	reason := "Not implemented"
	if len(message) > 0 {
		reason = message[0]
	}
	return Error{reason, http.StatusNotImplemented}
}

// InternalServerError returns a Error for a 500 Internal Server error.
func InternalServerError(message ...string) Error {
	reason := "Internal server error"
	if len(message) > 0 {
		reason = message[0]
	}
	return Error{reason, http.StatusInternalServerError}
}
