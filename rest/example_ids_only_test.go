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

import "strconv"

// MyResource represents a domain model for which we want to perform CRUD operations with.
// Endpoints can operate on any type of entity -- primitive, struct, or composite -- so long
// as it is serializable (by default, this means JSON-serializable via either MarshalJSON
// or JSON struct tags).
type MyResource struct {
	ID     int    `json:"id"`
	Foobar string `json:"foobar"`
}

// MyResourceHandler implements the ResourceHandler interface. It specifies the business
// logic for performing CRUD operations. BaseResourceHandler provides stubs for each method
// if you only need to implement certain operations (as this example illustrates).
type MyResourceHandler struct {
	BaseResourceHandler
}

// ResourceName is used to identify what resource a handler corresponds to and is used
// in the endpoint URLs, i.e. /api/:version/myresource.
func (m MyResourceHandler) ResourceName() string {
	return "myresource"
}

// ReadResourceList is the logic that corresponds to reading multiple resources, perhaps
// with specified query parameters accessed through the RequestContext. This is
// mapped to GET /api/:version/myresource. Typically, this would make some sort of database
// query to fetch the resources. It returns the slice of results, a cursor (or empty) string,
// and error (or nil). In this example, we illustrate how to use a query parameter to
// return only the IDs of our resources.
func (m MyResourceHandler) ReadResourceList(ctx RequestContext, limit int,
	cursor string, version string) ([]Resource, string, error) {
	// Make a database call here.
	resources := make([]Resource, 0, limit)
	resources = append(resources, &FooResource{ID: 1, Foobar: "hello"})
	resources = append(resources, &FooResource{ID: 2, Foobar: "world"})

	// ids_only is a query string parameter (i.e. /api/v1/myresource?ids_only=true).
	keysOnly, _ := strconv.ParseBool(ctx.ValueWithDefault("ids_only", "0").(string))

	if keysOnly {
		keys := make([]Resource, 0, len(resources))
		for _, resource := range resources {
			keys = append(keys, resource.(*FooResource).ID)
		}
		return keys, "", nil
	}

	return resources, "", nil
}

// This example shows how to implement a ResourceHandler which has an endpoint for
// fetching multiple resources and, optionally, retrieving only their IDs.
func Example_idsOnly() {
	api := NewAPI(NewConfiguration())

	// Call RegisterResourceHandler to wire up HelloWorldHandler.
	api.RegisterResourceHandler(MyResourceHandler{})

	// We're ready to hit our CRUD endpoints.
	api.Start(":8080")
}
