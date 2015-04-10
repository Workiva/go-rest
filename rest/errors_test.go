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
	assert.Equal(statusUnprocessableEntity, err.Status())

	err = UnauthorizedRequest("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(http.StatusUnauthorized, err.Status())

	err = MethodNotAllowed("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(http.StatusMethodNotAllowed, err.Status())

	err = ResourceNotPermitted("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(http.StatusForbidden, err.Status())

	err = InternalServerError("foo")
	assert.Equal("foo", err.Error())
	assert.Equal(http.StatusInternalServerError, err.Status())
}
