package examples

import (
	"fmt"
	"go-rest/server"
	"go-rest/server/context"
	"math/rand"
	"net/http"
	"strconv"
)

// FooResource represents a domain model for which we want to perform CRUD operations with.
// Endpoints can operate on any type of entity -- primitive, struct, or composite -- so long
// as it is serializable (by default, this means JSON-serializable via either MarshalJSON
// or JSON struct tags).
type FooResource struct {
	ID     int    `json:"id"`
	Foobar string `json:"foobar"`
}

// FooHandler implements the server.ResourceHandler interface. It specifies the business
// logic for performing CRUD operations.
type FooHandler struct{}

// ResourceName is used to identify what resource a handler corresponds to and is used
// in the endpoint URLs, i.e. /api/:version/foo.
func (f FooHandler) ResourceName() string {
	return "foo"
}

// CreateResource is the logic that corresponds to creating a new resource at
// POST /api/:version/foo. Typically, this would insert a record into a database.
// It returns the newly created resource or an error if the create failed.
func (f FooHandler) CreateResource(ctx context.RequestContext, data server.Payload,
	version string) (server.Resource, error) {
	// Make a database call here.
	id := rand.Int()
	created := &FooResource{ID: id, Foobar: data["foobar"].(string)}
	return created, nil
}

// ReadResource is the logic that corresponds to reading a single resource by its ID at
// GET /api/:version/foo/{id}. Typically, this would make some sort of database query to
// load the resource. If the resource doesn't exist, nil should be returned along with
// an appropriate error.
func (f FooHandler) ReadResource(ctx context.RequestContext, id string,
	version string) (server.Resource, error) {
	// Make a database call here.
	if id == "42" {
		return &FooResource{ID: 42, Foobar: "hello world"}, nil
	}
	return nil, server.ResourceNotFound(fmt.Sprintf("No resource with id %s", id))
}

// ReadResourceList is the logic that corresponds to reading multiple resources, perhaps
// with specified query parameters accessed through the context.RequestContext. This is
// mapped to GET /api/:version/foo. Typically, this would make some sort of database query
// to fetch the resources. It returns the slice of results, a cursor (or empty) string,
// and error (or nil).
func (f FooHandler) ReadResourceList(ctx context.RequestContext, limit int,
	cursor string, version string) ([]server.Resource, string, error) {
	// Make a database call here.
	resources := make([]server.Resource, 0, limit)
	resources = append(resources, &FooResource{ID: 1, Foobar: "hello"})
	resources = append(resources, &FooResource{ID: 2, Foobar: "world"})
	return resources, "", nil
}

// UpdateResource is the logic that corresponds to updating an existing resource at
// PUT /api/:version/foo/{id}. Typically, this would make some sort of database update
// call. It returns the updated resource or an error if the update failed.
func (f FooHandler) UpdateResource(ctx context.RequestContext, id string, data server.Payload,
	version string) (server.Resource, error) {
	// Make a database call here.
	updateId, _ := strconv.Atoi(id)
	foo := &FooResource{ID: updateId, Foobar: data["foobar"].(string)}
	return foo, nil
}

// DeleteResource is the logic that corresponds to deleting an existing resource at
// DELETE /api/:version/foo/{id}. Typically, this would make some sort of database
// delete call. It returns the deleted resource or an error if the delete failed.
func (f FooHandler) DeleteResource(ctx context.RequestContext, id string,
	version string) (server.Resource, error) {
	// Make a database call here.
	deleteId, _ := strconv.Atoi(id)
	foo := &FooResource{ID: deleteId, Foobar: "Goodbye world"}
	return foo, nil
}

// Authenticate is logic that is used to authenticate requests. The default behavior
// of Authenticate, seen in server.BaseResourceHandler, always returns nil, meaning
// all requests are authenticated. Returning an error means that the request is
// unauthorized and any error message will be sent back with the response.
func (f FooHandler) Authenticate(r http.Request) error {
	if secrets, ok := r.Header["Authorization"]; ok {
		if secrets[0] == "secret" {
			return nil
		}
	}

	return server.UnauthorizedRequest("You shall not pass")
}

// Start the REST server.
func simpleCrudMain() {
	api := server.NewRestAPI()

	// Call RegisterResourceHandler to wire up FooHandler.
	api.RegisterResourceHandler(FooHandler{})

	// We're ready to hit our CRUD endpoints.
	api.Start(":8080")
}
