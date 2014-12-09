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
	"errors"
	"fmt"
	"log"
	"net/http"
)

// MiddlewareResource represents a domain model for which we want to perform CRUD operations with.
// Endpoints can operate on any type of entity -- primitive, struct, or composite -- so long
// as it is serializable (by default, this means JSON-serializable via either MarshalJSON
// or JSON struct tags).
type MiddlewareResource struct {
	ID     int    `json:"id"`
	Foobar string `json:"foobar"`
}

// MiddlewareHandler implements the ResourceHandler interface. It specifies the business
// logic for performing CRUD operations. BaseResourceHandler provides stubs for
// each method if you only need to implement certain operations (as this example
// illustrates).
type MiddlewareHandler struct {
	BaseResourceHandler
}

// ResourceName is used to identify what resource a handler corresponds to and is used
// in the endpoint URLs, i.e. /api/:version/example.
func (e MiddlewareHandler) ResourceName() string {
	return "example"
}

// Authenticate is logic that is used to authenticate requests to this ResourceHandler.
// Returns nil if the request is authenticated or an error if it is not.
func (e MiddlewareHandler) Authenticate(r *http.Request) error {
	if r.Header.Get("Answer") != "42" {
		return errors.New("what is the answer?")
	}
	return nil
}

// ReadResource is the logic that corresponds to reading a single resource by its ID at
// GET /api/:version/example/{id}. Typically, this would make some sort of database query to
// load the resource. If the resource doesn't exist, nil should be returned along with
// an appropriate error.
func (e MiddlewareHandler) ReadResource(ctx RequestContext, id string, version string) (Resource, error) {
	// Make a database call here.
	if id == "42" {
		return &MiddlewareResource{ID: 42, Foobar: "hello world"}, nil
	}
	return nil, ResourceNotFound(fmt.Sprintf("No resource with id %s", id))
}

// ResourceHandler middleware is implemented as a closure which takes an http.HandlerFunc
// and returns one.
func HandlerMiddleware(wrapped http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s", r.URL.String())
		wrapped(w, r)
	}
}

// Global API middleware is implemented as a function which takes an http.ResponseWriter
// and http.Request and returns a bool indicating if the request should terminate or not.
func GlobalMiddleware(w http.ResponseWriter, r *http.Request) bool {
	log.Println(r)
	return false
}

// This example shows how to implement request middleware. ResourceHandlers provide the
// Authenticate method which is used to authenticate requests, but middleware allows
// you to insert additional authorization, logging, or other AOP-style operations.
func Example_middleware() {
	api := NewAPI(NewConfiguration())

	// Call RegisterResourceHandler to wire up MiddlewareHandler and apply middleware.
	api.RegisterResourceHandler(MiddlewareHandler{}, HandlerMiddleware)

	// Middleware provided to Start and StartTLS are invoked for every request handled
	// by the API.
	api.Start(":8080", GlobalMiddleware)
}
