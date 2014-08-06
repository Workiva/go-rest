package examples

import (
	"fmt"
	"go-rest/server"
	"go-rest/server/context"
)

// HelloWorldResource represents a domain model for which we want to perform CRUD operations with.
// Endpoints can operate on any type of entity -- primitive, struct, or composite -- so long
// as it is serializable (by default, this means JSON-serializable via either MarshalJSON
// or JSON struct tags).
type HelloWorldResource struct {
	ID     int    `json:"id"`
	Foobar string `json:"foobar"`
}

// HelloWorldHandler implements the server.ResourceHandler interface. It specifies the business
// logic for performing CRUD operations. server.BaseResourceHandler provides stubs for each method
// if you only need to implement certain operations (as this example illustrates).
type HelloWorldHandler struct {
	server.BaseResourceHandler
}

// ResourceName is used to identify what resource a handler corresponds to and is used
// in the endpoint URLs, i.e. /api/:version/foo.
func (h HelloWorldHandler) ResourceName() string {
	return "foo"
}

// ReadResource is the logic that corresponds to reading a single resource by its ID at
// GET /api/:version/foo/{id}. Typically, this would make some sort of database query to
// load the resource. If the resource doesn't exist, nil should be returned along with
// an appropriate error.
func (h HelloWorldHandler) ReadResource(ctx context.RequestContext, id string,
	version string) (server.Resource, error) {
	// Make a database call here.
	if id == "42" {
		return &HelloWorldResource{ID: 42, Foobar: "hello world"}, nil
	}
	return nil, server.ResourceNotFound(fmt.Sprintf("No resource with id %s", id))
}

// Start the REST server.
func helloWorldMain() {
	api := server.NewRestApi()

	// Call RegisterResourceHandler to wire up HelloWorldHandler.
	api.RegisterResourceHandler(HelloWorldHandler{})

	// We're ready to hit our CRUD endpoints.
	api.Start(":8080")
}
