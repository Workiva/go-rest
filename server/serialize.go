package server

import (
	"encoding/json"
	"fmt"
	"go-rest/server/context"
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

// SendResponse writes a success or error response to the provided http.ResponseWriter
// based on the contents of the context.RequestContext.
func SendResponse(w http.ResponseWriter, ctx context.RequestContext) {
	status := ctx.Status()
	requestError := ctx.Error()
	result := ctx.Result()

	serializer, err := responseSerializer(ctx.ResponseFormat())
	if err != nil {
		// Fall back to json serialization.
		serializer = jsonSerializer{}
		status = http.StatusNotImplemented
		requestError = err
	}

	if requestError != nil {
		if status < 400 {
			status = http.StatusInternalServerError
		}
		serializer.sendErrorResponse(w, requestError, status)
		return
	}

	nextURL, _ := ctx.NextURL()
	serializer.sendSuccessResponse(w, newSuccessResponse(result, nextURL), status)
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

var serializerRegistry map[string]ResponseSerializer = map[string]ResponseSerializer{"json": jsonSerializer{}}

// AddResponseSerializer registers the provided ResponseSerializer with the given format. If the
// format has already been registered, it will be overwritten.
func AddResponseSerializer(format string, serializer ResponseSerializer) {
	serializerRegistry[format] = serializer
}

// RemoveResponseSerializer unregisters the ResponseSerializer with the provided format. If the
// format hasn't been registered, this is a no-op.
func RemoveResponseSerializer(format string) {
	delete(serializerRegistry, format)
}

// responseSerializer returns a ResponseSerializer for the given format type. If the format
// is not implemented, the returned serializer will be nil and the error set.
func responseSerializer(format string) (ResponseSerializer, error) {
	if serializer, ok := serializerRegistry[format]; ok {
		return serializer, nil
	}
	return nil, fmt.Errorf("Format not implemented: %s", format)
}
