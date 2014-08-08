package rest

import (
	"encoding/json"
	"net/http"
)

// response is a data structure holding the serializable response body for a request and
// HTTP status code. It should be created using NewSuccessResponse or NewErrorResponse.
type response struct {
	Payload Payload
	Status  int
}

// ResponseSerializer is responsible for serializing REST responses and sending
// them back to the client.
type ResponseSerializer interface {

	// Serialize marshals a response payload into a byte slice to be sent over the wire.
	Serialize(Payload) ([]byte, error)

	// ContentType returns the MIME type of the response.
	ContentType() string
}

// jsonSerializer is an implementation of ResponseSerializer which serializes responses
// as JSON.
type jsonSerializer struct{}

// Serialize marshals a response payload into a JSON byte slice to be sent over the wire.
func (j jsonSerializer) Serialize(p Payload) ([]byte, error) {
	return json.Marshal(p)
}

// ContentType returns the JSON MIME type of the response.
func (j jsonSerializer) ContentType() string {
	return "application/json"
}

// NewSuccessResponse constructs a new response struct containing the Resource and,
// if provided, a "next" URL for retrieving the next page of results.
func NewSuccessResponse(r Resource, status int, nextURL string) response {
	payload := Payload{
		"success": true,
		"result":  r,
	}

	if nextURL != "" {
		payload["next"] = nextURL
	}

	response := response{
		Payload: payload,
		Status:  status,
	}

	return response
}

// NewErrorResponse constructs a new response struct containing the error message.
func NewErrorResponse(err error) response {
	payload := Payload{
		"success": false,
		"error":   err.Error(),
	}

	status := http.StatusInternalServerError
	if restError, ok := err.(Error); ok {
		status = restError.Status()
	}

	response := response{
		Payload: payload,
		Status:  status,
	}

	return response
}
