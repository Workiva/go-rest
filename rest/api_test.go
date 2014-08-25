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

// Ensures that the create handler returns a Not Implemented code if an invalid response
// format is provided.
func TestHandleCreateBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
	handler.On("CreateResource").Return(&TestResource{}, nil)

	api.RegisterResourceHandler(handler)
	createHandler, _ := api.(*muxAPI).getRouteHandler("foo:create")

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

// Ensures that the create handler returns an Internal Server Error code when the createFunc
// returns an error.
func TestHandleCreateBadCreate(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
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
		`{"error":"couldn't create","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the create handler returns the serialized resource and Created code when
// createFunc succeeds.
func TestHandleCreateHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
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
		`{"result":{"foo":"bar"},"success":true}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the create handler returns an Unauthorized code when the request is not
// authorized.
func TestHandleCreateNotAuthorized(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(fmt.Errorf("Not authorized"))

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

// Ensures that the read list handler returns a Not Implemented code if an invalid response
// format is provided.
func TestHandleReadListBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
	handler.On("ReadResourceList").Return([]Resource{}, "", nil)

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:readList")

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

// Ensures that the read list handler returns an Internal Server Error code when the readFunc returns an
// error.
func TestHandleReadListBadRead(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
	handler.On("ReadResourceList").Return(nil, "", fmt.Errorf("no resource"))

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:readList")

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
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
	handler.On("ReadResourceList").Return([]Resource{&TestResource{Foo: "hello"}}, "cursor123", nil)

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:readList")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusOK, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"next":"http://foo.com?next=cursor123","result":[{"foo":"hello"}],"success":true}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the read handler returns a Not Implemented code if an invalid response format is provided.
func TestHandleReadBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
	handler.On("ReadResource").Return(&TestResource{}, nil)

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

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
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
	handler.On("ReadResource").Return(nil, fmt.Errorf("no resource"))

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

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
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
	handler.On("ReadResource").Return(&TestResource{Foo: "hello"}, nil)

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

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

// Ensures that the update list handler returns a Not Implemented code if an invalid response format
// is provided.
func TestHandleUpdateListBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
	handler.On("UpdateResourceList").Return([]Resource{&TestResource{}}, nil)

	api.RegisterResourceHandler(handler)
	updateHandler, _ := api.(*muxAPI).getRouteHandler("foo:updateList")

	payload := []byte(`[{"foo": "bar"}]`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("PUT", "http://foo.com/api/v0.1/foo?format=blah", r)
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

// Ensures that the update list handler returns an Internal Server Error code when the
// updateListFunc returns an error.
func TestHandleUpdateListBadUpdate(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
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
		`{"error":"couldn't update","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that update list handler returns a Bad Request  code when the payload is
// not a list.
func TestHandleUpdateListPayloadNotList(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)

	api.RegisterResourceHandler(handler)
	updateHandler, _ := api.(*muxAPI).getRouteHandler("foo:updateList")

	payload := []byte(`{"foo": "bar"}`)
	r := bytes.NewReader(payload)
	req, _ := http.NewRequest("PUT", "http://foo.com/api/v0.1/foo", r)
	resp := httptest.NewRecorder()

	updateHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(http.StatusBadRequest, resp.Code, "Incorrect response code")
	assert.Equal(
		`{"error":"EOF","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update list handler returns the serialized resource and OK code when
// updateFunc succeeds.
func TestHandleUpdateListHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
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
		`{"result":[{"foo":"bar"}],"success":true}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update handler returns a Not Implemented code if an invalid response format is provided.
func TestHandleUpdateBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
	handler.On("UpdateResource").Return(&TestResource{}, nil)

	api.RegisterResourceHandler(handler)
	updateHandler, _ := api.(*muxAPI).getRouteHandler("foo:update")

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

// Ensures that the update handler returns an Internal Server Error code when the updateFunc returns an
// error.
func TestHandleUpdateBadUpdate(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
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
		`{"error":"couldn't update","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the update handler returns the serialized resource and OK code when updateFunc succeeds.
func TestHandleUpdateHappyPath(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
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
		`{"result":{"foo":"bar"},"success":true}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that the delete handler returns a Not Implemented code if an invalid response format is
// provided.
func TestHandleDeleteBadFormat(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
	handler.On("DeleteResource").Return(&TestResource{}, nil)

	api.RegisterResourceHandler(handler)
	deleteHandler, _ := api.(*muxAPI).getRouteHandler("foo:delete")

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

// Ensures that the delete handler returns an Internal Server Error code when the deleteFunc returns an
// error.
func TestHandleDeleteBadDelete(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
	handler.On("DeleteResource").Return(nil, fmt.Errorf("no resource"))

	api.RegisterResourceHandler(handler)
	deleteHandler, _ := api.(*muxAPI).getRouteHandler("foo:delete")

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
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
	handler.On("DeleteResource").Return(&TestResource{Foo: "hello"}, nil)

	api.RegisterResourceHandler(handler)
	deleteHandler, _ := api.(*muxAPI).getRouteHandler("foo:delete")

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
	api := NewAPI()

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(nil)
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
		`{"result":{"foo":"hello"},"success":true}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that outbound rules are applied.
func TestOutboundRules(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()
	rule := &Rule{
		Field:      "Foo",
		FieldAlias: "f",
		OutputOnly: true,
	}

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(NewRules((*TestResource)(nil), rule))
	handler.On("ReadResource").Return(&TestResource{Foo: "hello"}, nil)

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(
		`{"result":{"f":"hello"},"success":true}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that outbound rules are not applied if an error is returned by handler.
func TestOutboundRulesDontApplyOnError(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()
	rule := &Rule{
		Field:      "Foo",
		FieldAlias: "f",
		OutputOnly: true,
	}

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(NewRules((*TestResource)(nil), rule))
	handler.On("ReadResource").Return(nil, fmt.Errorf("oh snap"))

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(
		`{"error":"oh snap","success":false}`,
		resp.Body.String(),
		"Incorrect response string",
	)
}

// Ensures that outbound rules are not applied if a nil resource is returned by
// handler.
func TestOutboundRulesDontApplyOnNilResource(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	api := NewAPI()
	rule := &Rule{
		Field:      "Foo",
		FieldAlias: "f",
		OutputOnly: true,
	}

	handler.On("ResourceName").Return("foo")
	handler.On("Authenticate").Return(nil)
	handler.On("Rules").Return(NewRules((*TestResource)(nil), rule))
	handler.On("ReadResource").Return(nil, nil)

	api.RegisterResourceHandler(handler)
	readHandler, _ := api.(*muxAPI).getRouteHandler("foo:read")

	req, _ := http.NewRequest("GET", "http://foo.com/api/v0.1/foo/1", nil)
	resp := httptest.NewRecorder()

	readHandler.ServeHTTP(resp, req)

	handler.Mock.AssertExpectations(t)
	assert.Equal(
		`{"result":null,"success":true}`,
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
	api := NewAPI()

	assert.Equal([]string{"json"}, api.AvailableFormats())

	api.RegisterResponseSerializer("foo", &TestResponseSerializer{})

	assert.Equal([]string{"foo", "json"}, api.AvailableFormats())

	api.UnregisterResponseSerializer("foo")

	assert.Equal([]string{"json"}, api.AvailableFormats())
}

// Ensures that validateRules panics when the resource doesn't have a Rule field.
func TestValidateRulesBadField(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI()
	handler := new(MockResourceHandler)
	handler.On("ResourceName").Return("foo")
	handler.On("Rules").Return(NewRules((*TestResource)(nil), &Rule{Field: "bar"}))
	api.RegisterResourceHandler(handler)

	defer func() {
		r := recover()
		assert.NotNil(r, "Should have panicked")
	}()
	api.(*muxAPI).validateRules()
}

// Ensures that validateRules panics when a Rule has an incorrect type.
func TestValidateRulesBadType(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI()
	handler := new(MockResourceHandler)
	handler.On("ResourceName").Return("foo")
	handler.On("Rules").Return(NewRules((*TestResource)(nil), &Rule{Field: "Foo", Type: Int}))
	api.RegisterResourceHandler(handler)

	defer func() {
		r := recover()
		assert.NotNil(r, "Should have panicked")
	}()
	api.(*muxAPI).validateRules()
}

// Ensures that validateRules doesn't panic when the Rules are valid.
func TestValidateRulesHappyPath(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI()
	handler := new(MockResourceHandler)
	handler.On("ResourceName").Return("foo")
	handler.On("Rules").Return(NewRules((*TestResource)(nil), &Rule{
		Field: "Foo",
		Type:  String,
	}))
	api.RegisterResourceHandler(handler)

	defer func() {
		r := recover()
		assert.Nil(r, "Should not have panicked")
	}()
	api.(*muxAPI).validateRules()
}

// Ensures that validateRules doesn't panic when there are no Rules.
func TestValidateRulesNoRules(t *testing.T) {
	assert := assert.New(t)
	api := NewAPI()
	handler := new(MockResourceHandler)
	handler.On("ResourceName").Return("foo")
	handler.On("Rules").Return(nil)
	api.RegisterResourceHandler(handler)

	defer func() {
		r := recover()
		assert.Nil(r, "Should not have panicked")
	}()
	api.(*muxAPI).validateRules()
}
