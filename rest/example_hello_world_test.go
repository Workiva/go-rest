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

import "fmt"

// HelloWorldResource represents a domain model for which we want to perform CRUD operations with.
// Endpoints can operate on any type of entity -- primitive, struct, or composite -- so long
// as it is serializable (by default, this means JSON-serializable via either MarshalJSON
// or JSON struct tags).
type HelloWorldResource struct {
	ID     int    `json:"id"`
	Foobar string `json:"foobar"`
}

// HelloWorldHandler implements the ResourceHandler interface. It specifies the business
// logic for performing CRUD operations. BaseResourceHandler provides stubs for each method
// if you only need to implement certain operations (as this example illustrates).
type HelloWorldHandler struct {
	BaseResourceHandler
}

// ResourceName is used to identify what resource a handler corresponds to and is used
// in the endpoint URLs, i.e. /api/:version/helloworld.
func (h HelloWorldHandler) ResourceName() string {
	return "helloworld"
}

// ReadResource is the logic that corresponds to reading a single resource by its ID at
// GET /api/:version/helloworld/{id}. Typically, this would make some sort of database query to
// load the resource. If the resource doesn't exist, nil should be returned along with an
// appropriate error.
func (h HelloWorldHandler) ReadResource(ctx RequestContext, id string,
	version string) (Resource, error) {
	// Make a database call here.
	if id == "42" {
		return &HelloWorldResource{ID: 42, Foobar: "hello world"}, nil
	}
	return nil, ResourceNotFound(fmt.Sprintf("No resource with id %s", id))
}

// This example shows a minimal implementation of a ResourceHandler by using the
// BaseResourceHandler. It only implements an endpoint for fetching a resource.
func Example_helloWorld() {
	api := NewAPI(NewConfiguration())

	// Call RegisterResourceHandler to wire up HelloWorldHandler.
	api.RegisterResourceHandler(HelloWorldHandler{})

	// We're ready to hit our CRUD endpoints.
	api.Start(":8080")
}
