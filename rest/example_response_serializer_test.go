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

import "gopkg.in/yaml.v1"

// YAMLSerializer implements the ResponseSerializer interface.
type YAMLSerializer struct{}

// Serialize marshals a response payload into a byte slice to be sent over the wire.
func (y YAMLSerializer) Serialize(p Payload) ([]byte, error) {
	return yaml.Marshal(p)
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
	api := NewAPI(NewConfiguration())

	// Call RegisterResponseSerializer to wire up YAMLSerializer.
	api.RegisterResponseSerializer("yaml", YAMLSerializer{})

	// Call RegisterResourceHandler to wire up HelloWorldHandler.
	api.RegisterResourceHandler(HelloWorldHandler{})

	// We're ready to hit our CRUD endpoints. Use ?format=yaml to get responses as YAML.
	api.Start(":8080")
}
