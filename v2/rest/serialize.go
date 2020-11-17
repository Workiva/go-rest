/*
Copyright 2014 - 2015 Workiva, LLC

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

import (
	"encoding/json"
	"net/http"
	"reflect"
)

const (
	status   = "status"
	reason   = "reason"
	messages = "messages"
	result   = "result"
	results  = "results"
	next     = "next"
)

// response is a data structure holding the serializable response body for a request and
// HTTP status code. It should be created using NewResponse.
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

// NewResponse constructs a new response struct containing the payload to send back.
// It will either be a success or error response depending on the RequestContext.
func NewResponse(ctx RequestContext) response {
	err := ctx.Error()
	if err != nil {
		return newErrorResponse(ctx)
	}

	return newSuccessResponse(ctx)
}

// newSuccessResponse constructs a new response struct containing a resource response.
func newSuccessResponse(ctx RequestContext) response {
	r := ctx.Result()
	resultKey := result
	if r != nil && reflect.TypeOf(r).Kind() == reflect.Slice {
		resultKey = results
	}

	s := ctx.Status()
	response := response{Status: s}

	if s != http.StatusNoContent {
		payload := Payload{
			status:    s,
			reason:    http.StatusText(s),
			messages:  ctx.Messages(),
			resultKey: r,
		}

		if nextURL, err := ctx.NextURL(); err == nil && nextURL != "" {
			payload[next] = nextURL
		}

		response.Payload = payload
	}

	return response
}

// newErrorResponse constructs a new response struct containing an error message.
func newErrorResponse(ctx RequestContext) response {
	err := ctx.Error()
	s := http.StatusInternalServerError
	if restError, ok := err.(Error); ok {
		s = restError.Status()
	}

	payload := Payload{
		status:   s,
		reason:   http.StatusText(s),
		messages: ctx.Messages(),
	}

	response := response{
		Payload: payload,
		Status:  s,
	}

	return response
}
