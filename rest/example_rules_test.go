package rest

import "fmt"

// HelloWorldRuleHandler implements the ResourceHandler interface. It specifies the business
// logic for performing CRUD operations. BaseResourceHandler provides stubs for each method
// if you only need to implement certain operations (as this example illustrates).
type HelloWorldRuleHandler struct {
	BaseResourceHandler
}

// ResourceName is used to identify what resource a handler corresponds to and is used
// in the endpoint URLs, i.e. /api/:version/helloworld.
func (h HelloWorldRuleHandler) ResourceName() string {
	return "helloworld"
}

// ReadResource is the logic that corresponds to reading a single resource by its ID at
// GET /api/:version/helloworld/{id}. Typically, this would make some sort of database query to
// load the resource. If the resource doesn't exist, nil should be returned along with an
// appropriate error.
func (h HelloWorldRuleHandler) ReadResource(ctx RequestContext, id string,
	version string) (Resource, error) {
	// Make a database call here.
	if id == "42" {
		return &HelloWorldResource{ID: 42, Foobar: "hello world"}, nil
	}
	return nil, ResourceNotFound(fmt.Sprintf("No resource with id %s", id))
}

// Rules returns the resource rules to apply to incoming requests and outgoing responses. The
// default behavior, seen in BaseResourceHandler, is to apply no rules. In this example,
// different Rules are returned based on the version provided.
func (h HelloWorldRuleHandler) Rules(version string) []Rule {
	rules := []Rule{}
	if version == "1" {
		rules = append(rules, Rule{Field: "Foo", ValueName: "f"})
	} else if version == "2" {
		rules = append(rules, Rule{Field: "Foo", ValueName: "foo"})
	}
	return rules
}

// This example shows how Rules are used to provide fine-grained control over response
// output.
func Example_rules() {
	api := NewAPI()

	// Call RegisterResourceHandler to wire up HelloWorldRuleHandler.
	api.RegisterResourceHandler(HelloWorldRuleHandler{})

	// We're ready to hit our CRUD endpoints.
	api.Start(":8080")
}
