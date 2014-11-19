/*
Copyright 2014 Workiva, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
