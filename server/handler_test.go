package server

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-rest/server/context"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestResource struct {
	Foo string `json:"foo"`
}

type MockResourceHandler struct {
	mock.Mock
}

func (m *MockResourceHandler) ResourceName() string {
	args := m.Mock.Called()
	return args.String(0)
}

func (m *MockResourceHandler) CreateResource(r context.RequestContext, data Payload, version string) (Resource, error) {
	args := m.Mock.Called()
	resource := args.Get(0)
	if resource != nil {
		resource = resource.(*TestResource)
	}
	return resource, args.Error(1)
}

func (m *MockResourceHandler) ReadResource(r context.RequestContext, id string, version string) (Resource, error) {
	args := m.Mock.Called()
	resource := args.Get(0)
	if resource != nil {
		resource = resource.(*TestResource)
	}
	return resource, args.Error(1)
}

func (m *MockResourceHandler) ReadResourceList(r context.RequestContext, version string) ([]Resource, string, error) {
	args := m.Mock.Called()
	resources := args.Get(0)
	if resources != nil {
		return resources.([]Resource), args.String(1), args.Error(2)
	}
	return nil, args.String(1), args.Error(2)
}

func (m *MockResourceHandler) UpdateResource(r context.RequestContext, id string, data Payload, version string) (Resource, error) {
	args := m.Mock.Called()
	resource := args.Get(0)
	if resource != nil {
		resource = resource.(*TestResource)
	}
	return resource, args.Error(1)
}

func (m *MockResourceHandler) DeleteResource(r context.RequestContext, id string, version string) (Resource, error) {
	args := m.Mock.Called()
	resource := args.Get(0)
	if resource != nil {
		resource = resource.(*TestResource)
	}
	return resource, args.Error(1)
}

// Ensures that the create handler returns a Not Implemented code if an invalid response format is provided.
func TestHandleCreateBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("CreateResource").Return(&TestResource{}, nil)

	RegisterResourceHandler(router, handler)
	createHandler := router.Get("create").GetHandler()

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("POST", "http://foo.com/api/v0.1/foo?format=blah", r)
	resp := httptest.NewRecorder()

	createHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusNotImplemented, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"error":"Format not implemented: blah","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the create handler returns an Internal Server Error code when the createFunc returns an error.
func TestHandleCreateBadCreate(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("CreateResource").Return(nil, fmt.Errorf("couldn't create"))

	RegisterResourceHandler(router, handler)
	createHandler := router.Get("create").GetHandler()

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("POST", "http://foo.com/api/v0.1/foo", r)
	resp := httptest.NewRecorder()

	createHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusInternalServerError, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"error":"couldn't create","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the create handler returns the serialized resource and Created code when createFunc succeeds.
func TestHandleCreateHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("CreateResource").Return(&TestResource{Foo: "bar"}, nil)

	RegisterResourceHandler(router, handler)
	createHandler := router.Get("create").GetHandler()

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("POST", "http://foo.com/api/v0.1/foo", r)
	resp := httptest.NewRecorder()

	createHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusCreated, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"result":{"foo":"bar"},"success":true}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the read list handler returns a Not Implemented code if an invalid response format is provided.
func TestHandleReadListBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("ReadResourceList").Return([]Resource{}, "", nil)

	RegisterResourceHandler(router, handler)
	readHandler := router.Get("readList").GetHandler()

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo?format=blah", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusNotImplemented, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"error":"Format not implemented: blah","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the read list handler returns an Internal Server Error code when the readFunc returns an error.
func TestHandleReadListBadRead(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("ReadResourceList").Return(nil, "", fmt.Errorf("no resource"))

	RegisterResourceHandler(router, handler)
	readHandler := router.Get("readList").GetHandler()

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusInternalServerError, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"error":"no resource","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the read list handler returns the serialized resource and OK code when readFunc succeeds.
func TestHandleReadListHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("ReadResourceList").Return([]Resource{&TestResource{Foo: "hello"}}, "cursor123", nil)

	RegisterResourceHandler(router, handler)
	readHandler := router.Get("readList").GetHandler()

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusOK, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"next":"cursor123","result":[{"foo":"hello"}],"success":true}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the read handler returns a Not Implemented code if an invalid response format is provided.
func TestHandleReadBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("ReadResource").Return(&TestResource{}, nil)

	RegisterResourceHandler(router, handler)
	readHandler := router.Get("read").GetHandler()

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1?format=blah", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusNotImplemented, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"error":"Format not implemented: blah","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the read handler returns an Internal Server Error code when the readFunc returns an error.
func TestHandleReadBadRead(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("ReadResource").Return(nil, fmt.Errorf("no resource"))

	RegisterResourceHandler(router, handler)
	readHandler := router.Get("read").GetHandler()

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusInternalServerError, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"error":"no resource","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the read handler returns the serialized resource and OK code when readFunc succeeds.
func TestHandleReadHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("ReadResource").Return(&TestResource{Foo: "hello"}, nil)

	RegisterResourceHandler(router, handler)
	readHandler := router.Get("read").GetHandler()

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusOK, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"result":{"foo":"hello"},"success":true}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update handler returns a Not Implemented code if an invalid response format is provided.
func TestHandleUpdateBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("UpdateResource").Return(&TestResource{}, nil)

	RegisterResourceHandler(router, handler)
	updateHandler := router.Get("update").GetHandler()

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("PUT", "http://foo.com/api/v0.1/foo/1?format=blah", r)
	resp := httptest.NewRecorder()

	updateHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusNotImplemented, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"error":"Format not implemented: blah","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update handler returns an Internal Server Error code when the updateFunc returns an error.
func TestHandleUpdateBadUpdate(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("UpdateResource").Return(nil, fmt.Errorf("couldn't update"))

	RegisterResourceHandler(router, handler)
	updateHandler := router.Get("update").GetHandler()

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("PUT", "http://foo.com/api/v0.1/foo/1", r)
	resp := httptest.NewRecorder()

	updateHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusInternalServerError, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"error":"couldn't update","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update handler returns the serialized resource and OK code when updateFunc succeeds.
func TestHandleUpdateHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("UpdateResource").Return(&TestResource{Foo: "bar"}, nil)

	RegisterResourceHandler(router, handler)
	updateHandler := router.Get("update").GetHandler()

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("PUT", "http://foo.com/api/v0.1/foo/1", r)
	resp := httptest.NewRecorder()

	updateHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusOK, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"result":{"foo":"bar"},"success":true}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the delete handler returns a Not Implemented code if an invalid response format is provided.
func TestHandleDeleteBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("DeleteResource").Return(&TestResource{}, nil)

	RegisterResourceHandler(router, handler)
	deleteHandler := router.Get("delete").GetHandler()

	req, _ := http.NewRequest("DELETE", "http://foo.com/api/v0.1/foo/1?format=blah", nil)
	resp := httptest.NewRecorder()

	deleteHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusNotImplemented, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"error":"Format not implemented: blah","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the delete handler returns an Internal Server Error code when the deleteFunc returns an error.
func TestHandleDeleteBadDelete(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("DeleteResource").Return(nil, fmt.Errorf("no resource"))

	RegisterResourceHandler(router, handler)
	deleteHandler := router.Get("delete").GetHandler()

	req, _ := http.NewRequest("DELETE", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	deleteHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusInternalServerError, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"error":"no resource","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the delete handler returns the serialized resource and OK code when deleteFunc succeeds.
func TestHandleDeleteHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("DeleteResource").Return(&TestResource{Foo: "hello"}, nil)

	RegisterResourceHandler(router, handler)
	deleteHandler := router.Get("delete").GetHandler()

	req, _ := http.NewRequest("DELETE", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	deleteHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusOK, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"result":{"foo":"hello"},"success":true}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

func getMiddleware(called *bool) RequestMiddleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			*called = true
			h(w, r)
		}
	}
}

// Ensures that the middleware passed to RegisterResourceHandler is invoked during requests.
func TestApplyMiddleware(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("ReadResource").Return(&TestResource{Foo: "hello"}, nil)

	called := false
	RegisterResourceHandler(router, handler, getMiddleware(&called))
	readHandler := router.Get("read").GetHandler()

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.True(called, "Middleware was not invoked")
}
