package examples

import (
	"fmt"
	"go-rest/server"
	"go-rest/server/context"
	"log"
	"net/http"
)

// ExampleResource represents a domain model for which we want to perform CRUD operations with.
// Endpoints can operate on any type of entity -- primitive, struct, or composite -- so long
// as it is serializable (by default, this means JSON-serializable via either MarshalJSON
// or JSON struct tags).
type ExampleResource struct {
	ID     int    `json:"id"`
	Foobar string `json:"foobar"`
}

// ExampleHandler implements the server.ResourceHandler interface. It specifies the business
// logic for performing CRUD operations. server.BaseResourceHandler provides stubs for
// each method if you only need to implement certain operations (as this example
// illustrates).
type ExampleHandler struct {
	server.BaseResourceHandler
}

// ResourceName is used to identify what resource a handler corresponds to and is used
// in the endpoint URLs, i.e. /api/:version/foo.
func (e ExampleHandler) ResourceName() string {
	return "foo"
}

// ReadResource is the logic that corresponds to reading a single resource by its ID at
// GET /api/:version/foo/{id}. Typically, this would make some sort of database query to
// load the resource. If the resource doesn't exist, nil should be returned along with
// an appropriate error.
func (e ExampleHandler) ReadResource(ctx context.RequestContext, id string, version string) (server.Resource, error) {
	// Make a database call here.
	if id == "42" {
		return &ExampleResource{ID: 42, Foobar: "hello world"}, nil
	}
	return nil, server.ResourceNotFound(fmt.Sprintf("No resource with id %s", id))
}

// Middleware is implemented as a closure which takes an http.HandlerFunc and returns
// one.
func ExampleMiddleware(wrapped http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s", r.URL.String())
		wrapped(w, r)
	}
}

// Start the REST server.
func middlewareMain() {
	api := server.NewRestApi()

	// Call RegisterResourceHandler to wire up ExampleHandler and apply middleware.
	api.RegisterResourceHandler(ExampleHandler{}, ExampleMiddleware)

	// We're ready to hit our CRUD endpoints.
	api.Start(":8080")
}
