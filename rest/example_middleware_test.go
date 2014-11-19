/*
Copyright 2014 Workiva, LLC

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
	"fmt"
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

// ExampleHandler implements the ResourceHandler interface. It specifies the business
// logic for performing CRUD operations. BaseResourceHandler provides stubs for
// each method if you only need to implement certain operations (as this example
// illustrates).
type ExampleHandler struct {
	BaseResourceHandler
}

// ResourceName is used to identify what resource a handler corresponds to and is used
// in the endpoint URLs, i.e. /api/:version/example.
func (e ExampleHandler) ResourceName() string {
	return "example"
}

// ReadResource is the logic that corresponds to reading a single resource by its ID at
// GET /api/:version/example/{id}. Typically, this would make some sort of database query to
// load the resource. If the resource doesn't exist, nil should be returned along with
// an appropriate error.
func (e ExampleHandler) ReadResource(ctx RequestContext, id string, version string) (Resource, error) {
	// Make a database call here.
	if id == "42" {
		return &ExampleResource{ID: 42, Foobar: "hello world"}, nil
	}
	return nil, ResourceNotFound(fmt.Sprintf("No resource with id %s", id))
}

// Middleware is implemented as a closure which takes an http.HandlerFunc and returns
// one.
func ExampleMiddleware(wrapped http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s", r.URL.String())
		wrapped(w, r)
	}
}

// This example shows how to implement request middleware. ResourceHandlers provide the
// Authenticate method which is used to authenticate requests, but middleware allows
// you to insert additional authorization, logging, or other AOP-style operations.
func Example_middleware() {
	api := NewAPI(NewConfiguration())

	// Call RegisterResourceHandler to wire up ExampleHandler and apply middleware.
	api.RegisterResourceHandler(ExampleHandler{}, ExampleMiddleware)

	// We're ready to hit our CRUD endpoints.
	api.Start(":8080")
}
