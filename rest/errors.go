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
func ResourceNotFound(reason string) Error {
	return Error{reason, http.StatusNotFound}
}

// ResourceNotPermitted returns a Error for a 403 Forbidden error.
func ResourceNotPermitted(reason string) Error {
	return Error{reason, http.StatusForbidden}
}

// ResourceConflict returns a Error for a 409 Conflict error.
func ResourceConflict(reason string) Error {
	return Error{reason, http.StatusConflict}
}

// BadRequest returns a Error for a 400 Bad Request error.
func BadRequest(reason string) Error {
	return Error{reason, http.StatusBadRequest}
}

// UnprocessableRequest returns a Error for a 422 Unprocessable Entity error.
func UnprocessableRequest(reason string) Error {
	return Error{reason, 422}
}

// UnauthorizedRequest returns a Error for a 401 Unauthorized error.
func UnauthorizedRequest(reason string) Error {
	return Error{reason, http.StatusUnauthorized}
}

// NotImplemented returns a Error for a 501 Not Implemented error.
func NotImplemented(reason string) Error {
	return Error{reason, http.StatusNotImplemented}
}

// InternalServerError returns a Error for a 500 Internal Server error.
func InternalServerError(reason string) Error {
	return Error{reason, http.StatusInternalServerError}
}
