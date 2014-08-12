package rest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensures that the Error factories produce errors with expected values.
func TestErrors(t *testing.T) {
	assert := assert.New(t)

	err := ResourceNotFound("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(http.StatusNotFound, err.Status())

	err = ResourceNotPermitted("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(http.StatusForbidden, err.Status())

	err = ResourceConflict("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(http.StatusConflict, err.Status())

	err = BadRequest("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(http.StatusBadRequest, err.Status())

	err = UnprocessableRequest("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(422, err.Status())

	err = UnauthorizedRequest("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(http.StatusUnauthorized, err.Status())

	err = NotImplemented("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(http.StatusNotImplemented, err.Status())

	err = ResourceNotPermitted("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(http.StatusForbidden, err.Status())

	err = InternalServerError("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(http.StatusInternalServerError, err.Status())
}
