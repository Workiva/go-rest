package rest

import (
	"net/http"

	"gopkg.in/yaml.v1"
)

// YAMLSerializer implements the ResponseSerializer interface.
type YAMLSerializer struct{}

// SendErrorResponse writes an HTTP error response as YAML.
func (x YAMLSerializer) SendErrorResponse(w http.ResponseWriter, err error, errorCode int) {
	response := NewErrorResponse(err)
	yamlResponse, err := yaml.Marshal(response)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "text/yaml")
	w.WriteHeader(errorCode)
	w.Write(yamlResponse)
}

// SendSuccessResponse writes an HTTP success response as YAML.
func (x YAMLSerializer) SendSuccessResponse(w http.ResponseWriter, r Response, status int) {
	yamlResponse, err := yaml.Marshal(r)
	if err != nil {
		x.SendErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/yaml")
	w.WriteHeader(status)
	w.Write(yamlResponse)
}

// Start the REST server.
func Example_responseSerializer() {
	api := NewAPI()

	// Call RegisterResponseSerializer to wire up YAMLSerializer.
	api.RegisterResponseSerializer("yaml", YAMLSerializer{})

	// Call RegisterResourceHandler to wire up HelloWorldHandler.
	api.RegisterResourceHandler(HelloWorldHandler{})

	// We're ready to hit our CRUD endpoints. Use ?format=yaml to get responses as YAML.
	api.Start(":8080")
}
