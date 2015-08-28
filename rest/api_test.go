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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestResource struct {
	Foo string `json:"foo"`
}

type TestResourceSlice struct {
	Foo []string `json:"foo"`
}

type MockResourceHandler struct {
	mock.Mock
	BaseResourceHandler
}

func (m *MockResourceHandler) ResourceName() string {
	args := m.Mock.Called()
	return args.String(0)
}

func (m *MockResourceHandler) CreateResource(r RequestContext, data Payload,
	version string) (Resource, error) {
	args := m.Mock.Called()
	resource := args.Get(0)
	if resource != nil {
		resource = resource.(*TestResource)
	}
	return resource, args.Error(1)
}

func (m *MockResourceHandler) ReadResource(r RequestContext, id string,
	version string) (Resource, error) {
	args := m.Mock.Called()
	resource := args.Get(0)
	if resource != nil {
		resource = resource.(*TestResource)
	}
	return resource, args.Error(1)
}

func (m *MockResourceHandler) ReadResourceList(r RequestContext, limit int,
	cursor string, version string) ([]Resource, string, error) {
	args := m.Mock.Called()
	resources := args.Get(0)
	if resources != nil {
		return resources.([]Resource), args.String(1), args.Error(2)
	}
	return nil, args.String(1), args.Error(2)
}

func (m *MockResourceHandler) UpdateResourceList(r RequestContext, data []Payload,
	version string) ([]Resource, error) {
	args := m.Mock.Called()
	resource := args.Get(0)
	if resource != nil {
		return resource.([]Resource), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockResourceHandler) UpdateResource(r RequestContext, id string, data Payload,
	version string) (Resource, error) {
	args := m.Mock.Called()
	resource := args.Get(0)
	if resource != nil {
		resource = resource.(*TestResource)
	}
	return resource, args.Error(1)
}

func (m *MockResourceHandler) DeleteResource(r RequestContext, id string,
	version string) (Resource, error) {
	args := m.Mock.Called()
	resource := args.Get(0)
	if resource != nil {
		resource = resource.(*TestResource)
	}
	return resource, args.Error(1)
}

func (m *MockResourceHandler) Authenticate(r *http.Request) error {
	args := m.Mock.Called()
	return args.Error(0)
}

func (m *MockResourceHandler) ValidVersions() []string {
	args := m.Mock.Called()
	versions := args.Get(0)
	if versions != nil {
		return versions.([]string)
	}
	return nil
}

func (m *MockResourceHandler) Rules() Rules {
	args := m.Mock.Called()
	rules := args.Get(0)
	if rules != nil {
		return rules.(Rules)
	}
	return nil
}

// getRouteHandler returns the http.Handler for the API route with the given name.
// This is purely for testing purposes and shouldn't be used elsewhere.
func (r *muxAPI) getRouteHandler(name string) (http.Handler, error) {
	route := r.router.Get(name)
	if route == nil {
		return nil, fmt.Errorf("No API route with name %s", name)
	}

	return route.GetHandler(), nil
}

// Ensures that the create handler returns a Bad Request code if an invalid response
// format is provided.
func TestHandleCreateBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("CreateResource").Return(&TestResource{}, nil)

	api.RegisterResourceHandler(handler)
	createHandler, _ := api.(*muxAPI).getRouteHandler("foo:create")

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("POST", "http://foo.com/api/v0.1/foo?format=blah", r)
	resp := httptest.NewRecorder()

	createHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusBadRequest, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":["Format not implemented: blah"],"reason":"Bad Request","status":400}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the create handler returns an Internal Server Error code when the createFunc
// returns an error.
func TestHandleCreateBadCreate(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("CreateResource").Return(nil, fmt.Errorf("couldn't create"))

	api.RegisterResourceHandler(handler)
	createHandler, _ := api.(*muxAPI).getRouteHandler("foo:create")

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("POST", "http://foo.com/api/v0.1/foo", r)
	resp := httptest.NewRecorder()

	createHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusInternalServerError, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":["couldn't create"],"reason":"Internal Server Error","status":500}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the create handler returns the serialized resource and Created code when
// createFunc succeeds.
func TestHandleCreateHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("CreateResource").Return(&TestResource{Foo: "bar"}, nil)

	api.RegisterResourceHandler(handler)
	createHandler, _ := api.(*muxAPI).getRouteHandler("foo:create")

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("POST", "http://foo.com/api/v0.1/foo", r)
	resp := httptest.NewRecorder()

	createHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusCreated, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":[],"reason":"Created","result":{"foo":"bar"},"status":201}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the create handler returns No Content status code and
// an empty body when Resource is nil.
func TestHandleCreateNoContent(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("CreateResource").Return(nil, nil)

	api.RegisterResourceHandler(handler)
	createHandler, _ := api.(*muxAPI).getRouteHandler("foo:create")

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("POST", "http://foo.com/api/v0.1/foo", r)
	resp := httptest.NewRecorder()

	createHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusNoContent, resp.Code, "Incorrect response code")
	assert.Equal(
		"",
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the create handler returns an Unauthorized code when the request is not
// authorized.
func TestHandleCreateNotAuthorized(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(fmt.Errorf("Not authorized"))
	handler.On("ValidVersions").Return(nil)

	api.RegisterResourceHandler(handler)
	createHandler, _ := api.(*muxAPI).getRouteHandler("foo:create")

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("POST", "http://foo.com/api/v0.1/foo", r)
	resp := httptest.NewRecorder()

	createHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusUnauthorized, resp.Code, "Incorrect response code")
	assert.Equal("Not authorized", resp.Body.String(), "Incorrect response string")
}

// Ensures that the read list handler returns a Bad Request code if an invalid response
// format is provided.
func TestHandleReadListBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("ReadResourceList").Return([]Resource{}, "", nil)

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:readList")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo?format=blah", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusBadRequest, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":["Format not implemented: blah"],"reason":"Bad Request","status":400}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the read list handler returns an Internal Server Error code when the readFunc returns an
// error.
func TestHandleReadListBadRead(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("ReadResourceList").Return(nil, "", fmt.Errorf("no resource"))

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:readList")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusInternalServerError, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":["no resource"],"reason":"Internal Server Error","status":500}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the read list handler returns the serialized resource and OK code when readFunc succeeds.
func TestHandleReadListHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("ReadResourceList").Return([]Resource{&TestResource{Foo: "hello"}}, "cursor123", nil)

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:readList")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusOK, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":[],"next":"http://foo.com?next=cursor123","reason":"OK","results":[{"foo":"hello"}],"status":200}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the read handler returns a Bad Request code if an invalid response format is provided.
func TestHandleReadBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("ReadResource").Return(&TestResource{}, nil)

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1?format=blah", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusBadRequest, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":["Format not implemented: blah"],"reason":"Bad Request","status":400}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the read handler returns an Internal Server Error code when the readFunc returns an error.
func TestHandleReadBadRead(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("ReadResource").Return(nil, fmt.Errorf("no resource"))

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusInternalServerError, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":["no resource"],"reason":"Internal Server Error","status":500}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the read handler returns the serialized resource and OK code when readFunc succeeds.
func TestHandleReadHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("ReadResource").Return(&TestResource{Foo: "hello"}, nil)

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusOK, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":[],"reason":"OK","result":{"foo":"hello"},"status":200}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update list handler returns a Bad Request code if an invalid response format
// is provided.
func TestHandleUpdateListBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("UpdateResourceList").Return([]Resource{&TestResource{}}, nil)

	api.RegisterResourceHandler(handler)
	updateHandler, _ := api.(*muxAPI).getRouteHandler("foo:updateList")

	payload := []byte(`[{"foo": "bar"}]`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("PUT", "http://foo.com/api/v0.1/foo?format=blah", r)
	resp := httptest.NewRecorder()

	updateHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusBadRequest, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":["Format not implemented: blah"],"reason":"Bad Request","status":400}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update list handler returns an Internal Server Error code when the
// updateListFunc returns an error.
func TestHandleUpdateListBadUpdate(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("UpdateResourceList").Return(nil, fmt.Errorf("couldn't update"))

	api.RegisterResourceHandler(handler)
	updateHandler, _ := api.(*muxAPI).getRouteHandler("foo:updateList")

	payload := []byte(`[{"foo": "bar"}]`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("PUT", "http://foo.com/api/v0.1/foo", r)
	resp := httptest.NewRecorder()

	updateHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusInternalServerError, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":["couldn't update"],"reason":"Internal Server Error","status":500}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update list handler handles payloads that aren't lists.
func TestHandleUpdateListPayloadNotList(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("UpdateResourceList").Return([]Resource{&TestResource{Foo: "bar"}}, nil)

	api.RegisterResourceHandler(handler)
	updateHandler, _ := api.(*muxAPI).getRouteHandler("foo:updateList")

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("PUT", "http://foo.com/api/v0.1/foo", r)
	resp := httptest.NewRecorder()

	updateHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusOK, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":[],"reason":"OK","results":[{"foo":"bar"}],"status":200}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update list handler returns the serialized resource and OK code when
// updateFunc succeeds.
func TestHandleUpdateListHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("UpdateResourceList").Return([]Resource{&TestResource{Foo: "bar"}}, nil)

	api.RegisterResourceHandler(handler)
	updateHandler, _ := api.(*muxAPI).getRouteHandler("foo:updateList")

	payload := []byte(`[{"foo": "bar"}]`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("PUT", "http://foo.com/api/v0.1/foo", r)
	resp := httptest.NewRecorder()

	updateHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusOK, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":[],"reason":"OK","results":[{"foo":"bar"}],"status":200}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update handler returns a Bad Request code if an invalid response format is provided.
func TestHandleUpdateBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("UpdateResource").Return(&TestResource{}, nil)

	api.RegisterResourceHandler(handler)
	updateHandler, _ := api.(*muxAPI).getRouteHandler("foo:update")

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("PUT", "http://foo.com/api/v0.1/foo/1?format=blah", r)
	resp := httptest.NewRecorder()

	updateHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusBadRequest, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":["Format not implemented: blah"],"reason":"Bad Request","status":400}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update handler returns an Internal Server Error code when the updateFunc returns an
// error.
func TestHandleUpdateBadUpdate(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("UpdateResource").Return(nil, fmt.Errorf("couldn't update"))

	api.RegisterResourceHandler(handler)
	updateHandler, _ := api.(*muxAPI).getRouteHandler("foo:update")

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("PUT", "http://foo.com/api/v0.1/foo/1", r)
	resp := httptest.NewRecorder()

	updateHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusInternalServerError, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":["couldn't update"],"reason":"Internal Server Error","status":500}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update handler returns the serialized resource and OK code when updateFunc succeeds.
func TestHandleUpdateHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("UpdateResource").Return(&TestResource{Foo: "bar"}, nil)

	api.RegisterResourceHandler(handler)
	updateHandler, _ := api.(*muxAPI).getRouteHandler("foo:update")

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("PUT", "http://foo.com/api/v0.1/foo/1", r)
	resp := httptest.NewRecorder()

	updateHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusOK, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":[],"reason":"OK","result":{"foo":"bar"},"status":200}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the delete handler returns a Bad Request code if an invalid response format is
// provided.
func TestHandleDeleteBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("DeleteResource").Return(&TestResource{}, nil)

	api.RegisterResourceHandler(handler)
	deleteHandler, _ := api.(*muxAPI).getRouteHandler("foo:delete")

	req, _ := http.NewRequest("DELETE", "http://foo.com/api/v0.1/foo/1?format=blah", nil)
	resp := httptest.NewRecorder()

	deleteHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusBadRequest, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":["Format not implemented: blah"],"reason":"Bad Request","status":400}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the delete handler returns an Internal Server Error code when the deleteFunc returns an
// error.
func TestHandleDeleteBadDelete(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("DeleteResource").Return(nil, fmt.Errorf("no resource"))

	api.RegisterResourceHandler(handler)
	deleteHandler, _ := api.(*muxAPI).getRouteHandler("foo:delete")

	req, _ := http.NewRequest("DELETE", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	deleteHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusInternalServerError, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":["no resource"],"reason":"Internal Server Error","status":500}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the delete handler returns the serialized resource and OK code when deleteFunc succeeds.
func TestHandleDeleteHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("DeleteResource").Return(&TestResource{Foo: "hello"}, nil)

	api.RegisterResourceHandler(handler)
	deleteHandler, _ := api.(*muxAPI).getRouteHandler("foo:delete")

	req, _ := http.NewRequest("DELETE", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	deleteHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusOK, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"messages":[],"reason":"OK","result":{"foo":"hello"},"status":200}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

func getMiddleware(called *bool) RequestMiddleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*called = true
			h.ServeHTTP(w, r)
		})
	}
}

// Ensures that the middleware passed to RegisterResourceHandler is invoked during requests.
func TestApplyMiddleware(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	handler.On("ReadResource").Return(&TestResource{Foo: "hello"}, nil)

	called := false
	api.RegisterResourceHandler(handler, getMiddleware(&called))
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.True(called, "Middleware was not invoked")
	assert.Equal(
		`{"messages":[],"reason":"OK","result":{"foo":"hello"},"status":200}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that outbound rules are applied.
func TestOutboundRules(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})
	rule := &Rule{
		Field:      "Foo",
		FieldAlias: "f",
		OutputOnly: true,
	}

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(NewRules((*TestResource)(nil), rule))
	handler.On("ReadResource").Return(&TestResource{Foo: "hello"}, nil)

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(
		`{"messages":[],"reason":"OK","result":{"f":"hello"},"status":200}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that outbound rules are not applied if an error is returned by handler.
func TestOutboundRulesDontApplyOnError(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})
	rule := &Rule{
		Field:      "Foo",
		FieldAlias: "f",
		OutputOnly: true,
	}

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(NewRules((*TestResource)(nil), rule))
	handler.On("ReadResource").Return(nil, fmt.Errorf("oh snap"))

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(
		`{"messages":["oh snap"],"reason":"Internal Server Error","status":500}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that outbound rules are not applied if a nil resource is returned by
// handler.
func TestOutboundRulesDontApplyOnNilResource(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI(&Configuration{})
	rule := &Rule{
		Field:      "Foo",
		FieldAlias: "f",
		OutputOnly: true,
	}

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(NewRules((*TestResource)(nil), rule))
	handler.On("ReadResource").Return(nil, nil)

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(
		`{"messages":[],"reason":"OK","result":null,"status":200}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

type TestResponseSerializer struct{}

func (t TestResponseSerializer) Serialize(Payload) ([]byte, error) {
	return []byte{}, nil
}

func (t TestResponseSerializer) ContentType() string {
	return "application/foo"
}

// Ensures that RegisterResponseSerializer, UnregisterResponseSerializer, and
// AvailableFormats behave as expected.
func TestRegisterUnregisterResponseSerializer(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI(&Configuration{})

	assert.Equal([]string{"json"}, api.AvailableFormats())

	api.RegisterResponseSerializer("foo", &TestResponseSerializer{})

	assert.Equal([]string{"foo", "json"}, api.AvailableFormats())

	api.UnregisterResponseSerializer("foo")

	assert.Equal([]string{"json"}, api.AvailableFormats())
}

// Ensures that Validate returns an error when the resource doesn't have a Rule
// field.
func TestValidateBadField(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI(&Configuration{})
	handler := new(MockResourceHandler)
	handler.On("ResourceName").Return("foo")
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(NewRules((*TestResource)(nil), &Rule{Field: "bar"}))
	api.RegisterResourceHandler(handler)

	assert.Error(api.Validate())
}

// Ensures that Validate returns an error when a Rule has an incorrect type.
func TestValidateBadType(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI(&Configuration{})
	handler := new(MockResourceHandler)
	handler.On("ResourceName").Return("foo")
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(NewRules((*TestResource)(nil), &Rule{Field: "Foo", Type: Int}))
	api.RegisterResourceHandler(handler)

	assert.Error(api.Validate())
}

// Ensures that Validate returns nil when the Rules are valid.
func TestValidateHappyPath(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI(&Configuration{})
	handler := new(MockResourceHandler)
	handler.On("ResourceName").Return("foo")
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(NewRules((*TestResource)(nil), &Rule{
		Field: "Foo",
		Type:  String,
	}))
	api.RegisterResourceHandler(handler)

	assert.Nil(api.Validate())
}

// Ensures that Validate returns nil when there are no Rules.
func TestValidateNoRules(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI(&Configuration{})
	handler := new(MockResourceHandler)
	handler.On("ResourceName").Return("foo")
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	api.RegisterResourceHandler(handler)

	assert.Nil(api.Validate())
}

// Ensures that validateRulesOrPanic panics when the resource doesn't have a
// Rule field.
func TestValidateRulesOrPanicBadField(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI(&Configuration{})
	handler := new(MockResourceHandler)
	handler.On("ResourceName").Return("foo")
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(NewRules((*TestResource)(nil), &Rule{Field: "bar"}))
	api.RegisterResourceHandler(handler)

	defer func() {
		r := recover()
		assert.NotNil(r, "Should have panicked")
	}()
	api.(*muxAPI).validateRulesOrPanic()
}

// Ensures that validateRulesOrPanic panics when a Rule has an incorrect type.
func TestValidateRulesOrPanicBadType(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI(&Configuration{})
	handler := new(MockResourceHandler)
	handler.On("ResourceName").Return("foo")
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(NewRules((*TestResource)(nil), &Rule{Field: "Foo", Type: Int}))
	api.RegisterResourceHandler(handler)

	defer func() {
		r := recover()
		assert.NotNil(r, "Should have panicked")
	}()
	api.(*muxAPI).validateRulesOrPanic()
}

// Ensures that validateRulesOrPanic doesn't panic when the Rules are valid.
func TestValidateRulesOrPanicHappyPath(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI(&Configuration{})
	handler := new(MockResourceHandler)
	handler.On("ResourceName").Return("foo")
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(NewRules((*TestResource)(nil), &Rule{
		Field: "Foo",
		Type:  String,
	}))
	api.RegisterResourceHandler(handler)

	defer func() {
		r := recover()
		assert.Nil(r, "Should not have panicked")
	}()
	api.(*muxAPI).validateRulesOrPanic()
}

// Ensures that validateRulesOrPanic doesn't panic when there are no Rules.
func TestValidateRulesOrPanicNoRules(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI(&Configuration{})
	handler := new(MockResourceHandler)
	handler.On("ResourceName").Return("foo")
	handler.On("ValidVersions").Return(nil)
	handler.On("Rules").Return(&rules{})
	api.RegisterResourceHandler(handler)

	defer func() {
		r := recover()
		assert.Nil(r, "Should not have panicked")
	}()
	api.(*muxAPI).validateRulesOrPanic()
}

type httpHandler struct {
	called bool
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
}

// Ensures that middlewareProxy invokes middleware and doesn't delegate to the
// wrapped http.Handler if any middleware return true.
func TestMiddlewareProxyTerminate(t *testing.T) {
	assert := assert.New(t)
	handler := &httpHandler{}
	called := false
	middleware := func(w http.ResponseWriter, r *http.Request) bool {
		called = true
		return true
	}
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	proxy := wrapMiddleware(handler, middleware)

	proxy.ServeHTTP(w, req)

	assert.True(called)
	assert.False(handler.called)
}

// Ensures that middlewareProxy invokes middleware and delegates to the wrapped
// http.Handler if all middleware return false.
func TestMiddlewareProxyDelegate(t *testing.T) {
	assert := assert.New(t)
	handler := &httpHandler{}
	called := false
	middleware := func(w http.ResponseWriter, r *http.Request) bool {
		called = true
		return false
	}
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	proxy := wrapMiddleware(handler, middleware)

	proxy.ServeHTTP(w, req)

	assert.True(called)
	assert.True(handler.called)
}

// Ensure that version validation middleware passes through request
// on a valid version, and returns a 400 on an invalid version.
func TestVersionMiddleware(t *testing.T) {
	assert := assert.New(t)

	api := NewAPI(&Configuration{})
	handler := new(MockResourceHandler)
	handler.On("Authenticate").Return(nil)
	handler.On("ValidVersions").Return([]string{"1"})
	handler.On("Rules").Return(&rules{})
	handler.On("ResourceName").Return("widgets")
	handler.On("ReadResourceList").Return([]Resource{"foo"}, "", nil)

	api.RegisterResourceHandler(handler)

	// Valid version
	req, _ := http.NewRequest("GET", "http://example.com/api/v1/widgets", nil)
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	assert.Equal(w.Code, 200)
	assert.Contains(w.Body.String(), "foo")

	// Invalid version
	req, _ = http.NewRequest("GET", "http://example.com/api/v2/widgets", nil)
	w = httptest.NewRecorder()
	api.ServeHTTP(w, req)
	assert.Equal(w.Code, http.StatusBadRequest)
	assert.NotContains(w.Body.String(), "foo")
}
