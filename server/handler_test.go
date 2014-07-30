package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockResourceHandler struct {
	mock.Mock
}

func (m *MockResourceHandler) EndpointName() string {
	args := m.Mock.Called()
	return args.String(0)
}

func (m *MockResourceHandler) CreateResource(params *CreateParams) (interface{}, error) {
	args := m.Mock.Called()
	return args.Get(0), args.Error(0)
}

func (m *MockResourceHandler) ReadResource(params *ReadParams) (interface{}, error) {
	args := m.Mock.Called()
	return args.Get(0), args.Error(0)
}

func (m *MockResourceHandler) UpdateResource(params *UpdateParams) (interface{}, error) {
	args := m.Mock.Called()
	return args.Get(0), args.Error(0)
}

func (m *MockResourceHandler) DeleteResource(params *DeleteParams) (interface{}, error) {
	args := m.Mock.Called()
	return args.Get(0), args.Error(0)
}

type MockReader struct{}

func (r MockReader) Read(p []byte) (int, error) {
	return 0, nil
}

func TestHandleCreateBadFormat(t *testing.T) {
	//assert := assert.New(t)
	handler := new(MockResourceHandler)
	w := httptest.NewRecorder()

	handlerFunc := handleCreate(handler.CreateResource)
	r, _ := http.NewRequest("POST", "http://foo.com/api/v1/foos", &MockReader{})

	handlerFunc(w, r)

	fmt.Println(w.Code)
	fmt.Println(w.Body)
}
