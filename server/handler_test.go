package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-rest/server/context"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Resource struct {
	Foo string `json:"foo"`
}

type MockResourceHandler struct {
	mock.Mock
}

func (m *MockResourceHandler) ResourceName() string {
	args := m.Mock.Called()
	return args.String(0)
}

func (m *MockResourceHandler) CreateResource(r context.RequestContext, data map[string]interface{}) (interface{}, error) {
	args := m.Mock.Called()
	return args.Get(0).(*Resource), args.Error(1)
}

func (m *MockResourceHandler) ReadResource(r context.RequestContext, id string) (interface{}, error) {
	args := m.Mock.Called()
	resource := args.Get(0)
	if resource != nil {
		resource = resource.(*Resource)
	}
	return resource, args.Error(1)
}

func (m *MockResourceHandler) UpdateResource(r context.RequestContext, id string, data map[string]interface{}) (interface{}, error) {
	args := m.Mock.Called()
	return args.Get(0).(*Resource), args.Error(1)
}

func (m *MockResourceHandler) DeleteResource(r context.RequestContext, id string) (interface{}, error) {
	args := m.Mock.Called()
	return args.Get(0).(*Resource), args.Error(1)
}

// Ensures that the read handler returns a Not Implemented code if an invalid response format is provided.
func TestHandleReadBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")

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
		"Incorrect error string",
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
		"Incorrect error string",
	)
}

// Ensures that the read handler returns the serialized resource and OK code when readFunc succeeds.
func TestHandleReadHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	router := mux.NewRouter()

	handler.On("ResourceName").Return("foo")
	handler.On("ReadResource").Return(&Resource{Foo: "hello"}, nil)

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
		"Incorrect error string",
	)
}
