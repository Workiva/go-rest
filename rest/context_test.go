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
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensures that if a limit doesn't exist on the context, the default is returned.
func TestLimitDefault(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	ctx := NewContext(nil, req)
	assert.Equal(100, ctx.Limit())
}

// Ensures that if an invalid limit value is on the context, the default is returned.
func TestLimitBadValue(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	ctx := NewContext(nil, req)
	ctx = ctx.WithValue(limitKey, "blah")
	assert.Equal(100, ctx.Limit())
}

// Ensures that the correct limit is returned from the context.
func TestLimit(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	ctx := NewContext(nil, req)
	ctx = ctx.WithValue(limitKey, "5")
	assert.Equal(5, ctx.Limit())
}

// Ensures that Messages returns the messages set on the context.
func TestMessagesNoError(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	ctx := NewContext(nil, req)
	message := "foo"

	assert.Equal(0, len(ctx.Messages()))

	ctx.AddMessage(message)

	if assert.Equal(1, len(ctx.Messages())) {
		assert.Equal(message, ctx.Messages()[0])
	}
}

// Ensures that Messages returns the messages set on the context and the error message
// when an error is set.
func TestMessagesWithError(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	ctx := NewContext(nil, req)
	message := "foo"
	errMessage := "blah"
	err := fmt.Errorf(errMessage)

	ctx = ctx.setError(err)
	if assert.Equal(1, len(ctx.Messages())) {
		assert.Equal(errMessage, ctx.Messages()[0])
	}

	ctx.AddMessage(message)

	if assert.Equal(2, len(ctx.Messages())) {
		assert.Equal(message, ctx.Messages()[0])
		assert.Equal(errMessage, ctx.Messages()[1])
	}
}

// Ensures that Header returns the request Header.
func TestHeader(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	ctx := NewContext(nil, req)

	assert.Equal(req.Header, ctx.Header())
}
