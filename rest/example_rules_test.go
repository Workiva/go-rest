package rest

import (
	"fmt"
	"math/rand"
)

// ResourceWithSecret represents a domain model for which we want to perform CRUD operations
// with. Endpoints can operate on any type of entity -- primitive, struct, or composite -- so
// long as it is serializable (by default, this means JSON-serializable via either MarshalJSON
// or JSON struct tags). The resource in this example has a field, "Secret", which we don't
// want to include in REST responses.
type ResourceWithSecret struct {
	ID     int    `json:"id"`
	Foo    string `json:"foo"`
	Secret string
}

// ResourceWithSecretHandler implements the ResourceHandler interface. It specifies the
// business logic for performing CRUD operations. BaseResourceHandler provides stubs for each
// method if you only need to implement certain operations (as this example illustrates).
type ResourceWithSecretHandler struct {
	BaseResourceHandler
}

// ResourceName is used to identify what resource a handler corresponds to and is used
// in the endpoint URLs, i.e. /api/:version/resource.
func (r ResourceWithSecretHandler) ResourceName() string {
	return "resource"
}

// CreateResource is the logic that corresponds to creating a new resource at
// POST /api/:version/resource. Typically, this would insert a record into a database.
// It returns the newly created resource or an error if the create failed. Because our Rules
// specify types, we can access the Payload data in a type-safe way.
func (r ResourceWithSecretHandler) CreateResource(ctx RequestContext, data Payload,
	version string) (Resource, error) {
	// Make a database call here.
	id := rand.Int()
	foo, _ := data.GetString("foo")
	created := &ResourceWithSecret{ID: id, Foo: foo, Secret: "secret"}
	return created, nil
}

// ReadResource is the logic that corresponds to reading a single resource by its ID at
// GET /api/:version/resource/{id}. Typically, this would make some sort of database query to
// load the resource. If the resource doesn't exist, nil should be returned along with an
// appropriate error.
func (r ResourceWithSecretHandler) ReadResource(ctx RequestContext, id string,
	version string) (Resource, error) {
	// Make a database call here.
	if id == "42" {
		return &ResourceWithSecret{
			ID:     42,
			Foo:    "hello world",
			Secret: "keep it secret, keep it safe",
		}, nil
	}
	return nil, ResourceNotFound(fmt.Sprintf("No resource with id %s", id))
}

// Rules returns the resource rules to apply to incoming requests and outgoing responses. The
// default behavior, seen in BaseResourceHandler, is to apply no rules. In this example,
// different Rules are returned based on the version provided. Note that a Rule is not
// specified for the "Secret" field. This means that field will not be included in the
// response. The "Type" field on a Rule indicates the type the incoming data should be
// coerced to. If coercion fails, an error indicating this will be sent back in the response.
// If no type is specified, no coercion will be performed.
func (r ResourceWithSecretHandler) Rules(version string) []Rule {
	rules := []Rule{}
	if version == "1" {
		rules = append(rules,
			Rule{Field: "ID", FieldAlias: "id", Type: Int, OutputOnly: true},
			Rule{Field: "Foo", FieldAlias: "f", Type: String, Required: true},
		)
	} else if version == "2" {
		rules = append(rules,
			Rule{Field: "ID", FieldAlias: "id", Type: Int, OutputOnly: true},
			Rule{Field: "Foo", FieldAlias: "foo", Type: String, Required: true},
		)
	}
	return rules
}

// This example shows how Rules are used to provide fine-grained control over response
// output.
func Example_rules() {
	api := NewAPI()

	// Call RegisterResourceHandler to wire up ResourceWithSecretHandler.
	api.RegisterResourceHandler(ResourceWithSecretHandler{})

	// We're ready to hit our CRUD endpoints.
	api.Start(":8080")
}
