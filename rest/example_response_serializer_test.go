package rest

import "gopkg.in/yaml.v1"

// YAMLSerializer implements the ResponseSerializer interface.
type YAMLSerializer struct{}

// Serialize marshals a response payload into a byte slice to be sent over the wire.
func (y YAMLSerializer) Serialize(r map[string]interface{}) ([]byte, error) {
	return yaml.Marshal(r)
}

// ContentType returns the MIME type of the response.
func (y YAMLSerializer) ContentType() string {
	return "text/yaml"
}

// This example shows how to implement a custom ResponseSerializer. The format responses
// are sent in is specified by the "format" query string parameter. By default, json is
// the only available format, but the ResponseSerializer interface allows different
// formats to be implemented.
func Example_responseSerializer() {
	api := NewAPI()

	// Call RegisterResponseSerializer to wire up YAMLSerializer.
	api.RegisterResponseSerializer("yaml", YAMLSerializer{})

	// Call RegisterResourceHandler to wire up HelloWorldHandler.
	api.RegisterResourceHandler(HelloWorldHandler{})

	// We're ready to hit our CRUD endpoints. Use ?format=yaml to get responses as YAML.
	api.Start(":8080")
}
