/*
Copyright 2014 - 2015 Workiva, LLC

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
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Handlers
type TestResourceHandler struct {
	BaseResourceHandler
}

func (t TestResourceHandler) ResourceName() string {
	return "widgets"
}

func (t TestResourceHandler) CreateResource(r RequestContext, data Payload,
	version string) (Resource, error) {

	resource := map[string]string{"test": "resource"}
	return resource, nil
}

func (t TestResourceHandler) ReadResource(r RequestContext, id string,
	version string) (Resource, error) {

	resource := map[string]string{"test": "resource"}
	return resource, nil
}

type ComplexTestResourceHandler struct {
	BaseResourceHandler
}

type contextKey string

func (t ComplexTestResourceHandler) ResourceName() string {
	return "resources"
}

func (t ComplexTestResourceHandler) CreateURI() string {
	return "/api/v{version:[^/]+}/{company}/{category}/resources"
}
func (t ComplexTestResourceHandler) CreateResource(r RequestContext, data Payload,
	version string) (Resource, error) {

	resource := map[string]string{"test": "resource"}
	return resource, nil
}

// Ensures that we're correctly setting a value on the http.Request Context
func TestSetValueOnRequestContext(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	k := contextKey("mykey")
	val := req.Context().Value(k)
	assert.Nil(t, val)

	expected := "myval"
	req = setValueOnRequestContext(req, k, expected)
	actual := req.Context().Value(k)
	assert.Equal(t, expected, actual)
}

func TestValueReturnsRequest(t *testing.T) {
	assert := assert.New(t)
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	ctx := NewContext(req, nil)
	val := ctx.Value(requestKey)
	assert.Equal(req, val, "Value called with the requestKey should return the request")
}

func TestValueReturnsPreviouslySetValue(t *testing.T) {
	assert := assert.New(t)
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	ctx := NewContext(req, nil)
	k := contextKey("mykey")
	expected := "myval"
	ctx = ctx.WithValue(k, expected)

	actual := ctx.Value(k)
	assert.Equal(expected, actual, "Value called with a key should find the value on the request's context if it exists")
}

func TestValueReturnsNilIfNoKey(t *testing.T) {
	assert := assert.New(t)
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	ctx := NewContext(req, nil).(*requestContext)
	k := contextKey("mykey")

	actual := ctx.Value(k)
	assert.Nil(actual, "Value called with a non-existent key should return nil")
}

// Ensures that if a limit doesn't exist on the context, the default is returned.
func TestLimitDefault(t *testing.T) {
	assert := assert.New(t)
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	writer := httptest.NewRecorder()
	ctx := NewContext(req, writer)
	assert.Equal(100, ctx.Limit())
}

// Ensures that if an invalid limit value is on the context, the default is returned.
func TestLimitBadValue(t *testing.T) {
	assert := assert.New(t)
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	writer := httptest.NewRecorder()
	ctx := NewContext(req, writer)
	ctx = ctx.WithValue(limitKey, "blah")
	assert.Equal(100, ctx.Limit())
}

// Ensures that the correct limit is returned from the context.
func TestLimit(t *testing.T) {
	assert := assert.New(t)
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	writer := httptest.NewRecorder()
	ctx := NewContext(req, writer)
	ctx = ctx.WithValue(limitKey, "5")
	assert.Equal(5, ctx.Limit())
}

// Ensures that Messages returns the messages set on the context.
func TestMessagesNoError(t *testing.T) {
	assert := assert.New(t)
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	writer := httptest.NewRecorder()
	ctx := NewContext(req, writer)
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
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	writer := httptest.NewRecorder()
	ctx := NewContext(req, writer)
	message := "foo"
	errMessage := "blah"
	err = fmt.Errorf(errMessage)

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
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	writer := httptest.NewRecorder()
	ctx := NewContext(req, writer)

	assert.Equal(req.Header, ctx.Header())
}

// Ensures that Body returns a buffer containing the request Body.
func TestBody(t *testing.T) {
	assert := assert.New(t)
	payload := []byte(`[{"foo": "bar"}]`)
	r := bytes.NewReader(payload)
	req, err := http.NewRequest("GET", "http://example.com/foo", r)
	require.NoError(t, err)

	writer := httptest.NewRecorder()
	ctx := NewContext(req, writer)

	assert.Equal(payload, ctx.Body().Bytes())
}

func TestBuildURL(t *testing.T) {
	assert := assert.New(t)

	api := NewAPI(NewConfiguration())
	api.RegisterResourceHandler(TestResourceHandler{})
	api.RegisterResourceHandler(ComplexTestResourceHandler{})

	req, err := http.NewRequest("GET", "https://example.com/api/v1/widgets", nil)
	require.NoError(t, err)
	req = setValueOnRequestContext(req, "version", "1")

	writer := httptest.NewRecorder()
	ctx := NewContextWithRouter(req, writer, api.(*muxAPI).router)

	url, err := ctx.BuildURL("widgets", HandleCreate, nil)
	require.NoError(t, err)
	assert.Equal(url.String(), "http://example.com/api/v1/widgets")

	url, err = ctx.BuildURL("widgets", HandleRead, RouteVars{"resource_id": "111"})
	require.NoError(t, err)
	assert.Equal(url.String(), "http://example.com/api/v1/widgets/111")

	url, err = ctx.BuildURL("widgets", HandleRead, RouteVars{"resource_id": "111"})
	require.NoError(t, err)
	assert.Equal(url.Path, "/api/v1/widgets/111")

	// Secure request should produce https URL
	req.TLS = &tls.ConnectionState{}
	url, err = ctx.BuildURL("widgets", HandleRead, RouteVars{"resource_id": "222"})
	require.NoError(t, err)
	assert.Equal(url.String(), "https://example.com/api/v1/widgets/222")

	// Make sure this works with another version number
	ctx = ctx.WithValue("version", "2")
	url, err = ctx.BuildURL("resources", HandleCreate, RouteVars{
		"company":  "acme",
		"category": "anvils"})
	require.NoError(t, err)
	assert.Equal(url.String(), "https://example.com/api/v2/acme/anvils/resources")
}
