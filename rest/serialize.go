package rest

import (
	"encoding/json"
	"net/http"
)

// Response is a data structure holding the serializable response body for a request.
// It should be created using newSuccessResponse or newErrorResponse.
type Response map[string]interface{}

// ResponseSerializer is responsible for serializing REST responses and sending
// them back to the client.
type ResponseSerializer interface {

	// SendSuccessResponse writes a response containing the provided response body
	// and status code to the http.ResponseWriter. If the body is not serializable,
	// an error response will be written instead.
	SendSuccessResponse(http.ResponseWriter, Response, int)

	// SendErrorResponse writes a response containing an error message and code to
	// the provided http.ResponseWriter.
	SendErrorResponse(http.ResponseWriter, error, int)
}

// jsonSerializer is an implementation of ResponseSerializer which serializes responses
// as JSON.
type jsonSerializer struct{}

// SendErrorResponse writes a response containing an error message and code to the provided
// http.ResponseWriter.
func (j jsonSerializer) SendErrorResponse(w http.ResponseWriter, err error, errorCode int) {
	response := NewErrorResponse(err)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCode)
	w.Write(jsonResponse)
}

// SendSuccessResponse writes a response containing the provided response body and status
// code to the http.ResponseWriter. If the body is not serializable, an error response
// will be written instead.
func (j jsonSerializer) SendSuccessResponse(w http.ResponseWriter, r Response, status int) {
	jsonResponse, err := json.Marshal(r)
	if err != nil {
		j.SendErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonResponse)
}

// NewSuccessResponse constructs a new response struct containing the Resource and,
// if provided, a "next" URL for retrieving the next page of results.
func NewSuccessResponse(r Resource, nextURL string) Response {
	response := Response{
		"success": true,
		"result":  r,
	}

	if nextURL != "" {
		response["next"] = nextURL
	}

	return response
}

// NewErrorResponse constructs a new response struct containing the error message.
func NewErrorResponse(err error) Response {
	return Response{
		"success": false,
		"error":   err.Error(),
	}
}
