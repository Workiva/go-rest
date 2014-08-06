package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensures that the RestError factories produce errors with expected values.
func TestErrors(t *testing.T) {
	assert := assert.New(t)
	var err RestError

	err = ResourceNotFound()
	assert.Equal("Resource not found", err.Error())
	assert.Equal(http.StatusNotFound, err.Status())
	err = ResourceNotFound("foo")
	assert.Equal("foo", err.Error())

	err = ResourceNotPermitted()
	assert.Equal("Resource forbidden", err.Error())
	assert.Equal(http.StatusForbidden, err.Status())
	err = ResourceNotPermitted("foo")
	assert.Equal("foo", err.Error())

	err = ResourceConflict()
	assert.Equal("Resource conflict", err.Error())
	assert.Equal(http.StatusConflict, err.Status())
	err = ResourceConflict("foo")
	assert.Equal("foo", err.Error())

	err = BadRequest()
	assert.Equal("Bad request", err.Error())
	assert.Equal(http.StatusBadRequest, err.Status())
	err = BadRequest("foo")
	assert.Equal("foo", err.Error())

	err = UnprocessableRequest()
	assert.Equal("Unprocessable request", err.Error())
	assert.Equal(422, err.Status())
	err = UnprocessableRequest("foo")
	assert.Equal("foo", err.Error())

	err = UnauthorizedRequest()
	assert.Equal("Unauthorized request", err.Error())
	assert.Equal(http.StatusUnauthorized, err.Status())
	err = UnauthorizedRequest("foo")
	assert.Equal("foo", err.Error())

	err = NotImplemented()
	assert.Equal("Not implemented", err.Error())
	assert.Equal(http.StatusNotImplemented, err.Status())
	err = ResourceNotPermitted("foo")
	assert.Equal("foo", err.Error())

	err = InternalServerError()
	assert.Equal("Internal server error", err.Error())
	assert.Equal(http.StatusInternalServerError, err.Status())
	err = InternalServerError("foo")
	assert.Equal("foo", err.Error())
}
