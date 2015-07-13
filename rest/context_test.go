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
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	gContext "github.com/gorilla/context"
	"github.com/stretchr/testify/assert"
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

func TestBuildURL(t *testing.T) {
	assert := assert.New(t)

	api := NewAPI(NewConfiguration())
	api.RegisterResourceHandler(TestResourceHandler{})
	api.RegisterResourceHandler(ComplexTestResourceHandler{})

	req, _ := http.NewRequest("GET", "https://example.com/api/v1/widgets", nil)
	gContext.Set(req, "version", "1")

	ctx := NewContextWithRouter(nil, req, api.(*muxAPI).router)

	url, _ := ctx.BuildURL("widgets", HandleCreate, nil)
	assert.Equal(url, "http://example.com/api/v1/widgets")

	url, _ = ctx.BuildURL("widgets", HandleRead, RouteVars{"resource_id": "111"})
	assert.Equal(url, "http://example.com/api/v1/widgets/111")

	url, _ = ctx.BuildPath("widgets", HandleRead, RouteVars{"resource_id": "111"})
	assert.Equal(url, "/api/v1/widgets/111")

	// Secure request should produce https URL
	req.TLS = &tls.ConnectionState{}
	url, _ = ctx.BuildURL("widgets", HandleRead, RouteVars{"resource_id": "222"})
	assert.Equal(url, "https://example.com/api/v1/widgets/222")

	// Make sure this works with another version number
	gContext.Set(req, "version", "2")
	url, _ = ctx.BuildURL("resources", HandleCreate, RouteVars{
		"company":  "acme",
		"category": "anvils"})
	assert.Equal(url, "https://example.com/api/v2/acme/anvils/resources")
}
