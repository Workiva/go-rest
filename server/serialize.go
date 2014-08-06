package server

import (
	"encoding/json"
	"net/http"
)

// response is a data structure holding the serializable response body for a request.
// It should be created using newSuccessResponse or newErrorResponse.
type response map[string]interface{}

// ResponseSerializer is responsible for serializing REST responses and sending
// them back to the client.
type ResponseSerializer interface {
	sendSuccessResponse(http.ResponseWriter, response, int)
	sendErrorResponse(http.ResponseWriter, error, int)
}

// jsonSerializer is an implementation of ResponseSerializer which serializes responses
// as JSON.
type jsonSerializer struct{}

// sendErrorResponse writes a response containing an error message and code to the provided
// http.ResponseWriter.
func (j jsonSerializer) sendErrorResponse(w http.ResponseWriter, err error, errorCode int) {
	response := newErrorResponse(err)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCode)
	w.Write(jsonResponse)
}

// sendSuccessResponse writes a response containing the provided response body and status
// code to the http.ResponseWriter. If the body is not serializable, an error response
// will be written instead.
func (j jsonSerializer) sendSuccessResponse(w http.ResponseWriter, r response, status int) {
	jsonResponse, err := json.Marshal(r)
	if err != nil {
		j.sendErrorResponse(w, err, 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonResponse)
}

// newSuccessResponse constructs a new response struct containing the Resource and,
// if provided, a "next" URL for retrieving the next page of results.
func newSuccessResponse(r Resource, nextURL string) response {
	response := response{
		"success": true,
		"result":  r,
	}

	if nextURL != "" {
		response["next"] = nextURL
	}

	return response
}

// newErrorResponse constructs a new response struct containing the error message.
func newErrorResponse(err error) response {
	return response{
		"success": false,
		"error":   err.Error(),
	}
}
