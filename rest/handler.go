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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

// Resource represents a domain model.
type Resource interface{}

// ResourceHandler specifies the endpoint handlers for working with a resource. This
// consists of the business logic for performing CRUD operations.
type ResourceHandler interface {
	// ResourceName is used to identify what resource a handler corresponds to and is
	// used in the endpoint URLs, i.e. /api/:version/resourceName. This should be
	// unique across all ResourceHandlers.
	ResourceName() string

	// CreateURI returns the URI for creating a resource.
	CreateURI() string

	// CreateDocumentation returns a string describing the handler's create endpoint.
	CreateDocumentation() string

	// ReadURI returns the URI for reading a specific resource.
	ReadURI() string

	// ReadDocumentation returns a string describing the handler's read endpoint.
	ReadDocumentation() string

	// ReadListURI returns the URI for reading a list of resources.
	ReadListURI() string

	// ReadListDocumentation returns a string describing the handler's read list endpoint.
	ReadListDocumentation() string

	// UpdateURI returns the URI for updating a specific resource.
	UpdateURI() string

	// UpdateDocumentation returns a string describing the handler's update endpoint.
	UpdateDocumentation() string

	// UpdateListURI returns the URI for updating a list of resources.
	UpdateListURI() string

	// UpdateListDocumentation returns a string describing the handler's update list
	// endpoint.
	UpdateListDocumentation() string

	// DeleteURI returns the URI for deleting a specific resource.
	DeleteURI() string

	// DeleteDocumentation returns a string describing the handler's delete endpoint.
	DeleteDocumentation() string

	// CreateResource is the logic that corresponds to creating a new resource at
	// POST /api/:version/resourceName. Typically, this would insert a record into a
	// database. It returns the newly created resource or an error if the create failed.
	CreateResource(RequestContext, Payload, string) (Resource, error)

	// ReadResourceList is the logic that corresponds to reading multiple resources,
	// perhaps with specified query parameters accessed through the RequestContext. This
	// is mapped to GET /api/:version/resourceName. Typically, this would make some sort
	// of database query to fetch the resources. It returns the slice of results, a
	// cursor (or empty) string, and error (or nil).
	ReadResourceList(RequestContext, int, string, string) ([]Resource, string, error)

	// ReadResource is the logic that corresponds to reading a single resource by its ID
	// at GET /api/:version/resourceName/{id}. Typically, this would make some sort of
	// database query to load the resource. If the resource doesn't exist, nil should be
	// returned along with an appropriate error.
	ReadResource(RequestContext, string, string) (Resource, error)

	// UpdateResourceList is the logic that corresponds to updating a collection of
	// resources at PUT /api/:version/resourceName. Typically, this would make some
	// sort of database update call. It returns the updated resources or an error if
	// the update failed.
	UpdateResourceList(RequestContext, []Payload, string) ([]Resource, error)

	// UpdateResource is the logic that corresponds to updating an existing resource at
	// PUT /api/:version/resourceName/{id}. Typically, this would make some sort of
	// database update call. It returns the updated resource or an error if the update
	// failed.
	UpdateResource(RequestContext, string, Payload, string) (Resource, error)

	// DeleteResource is the logic that corresponds to deleting an existing resource at
	// DELETE /api/:version/resourceName/{id}. Typically, this would make some sort of
	// database delete call. It returns the deleted resource or an error if the delete
	// failed.
	DeleteResource(RequestContext, string, string) (Resource, error)

	// Authenticate is logic that is used to authenticate requests. The default behavior
	// of Authenticate, seen in BaseResourceHandler, always returns nil, meaning all
	// requests are authenticated. Returning an error means that the request is
	// unauthorized and any error message will be sent back with the response.
	Authenticate(*http.Request) error

	// Rules returns the resource rules to apply to incoming requests and outgoing
	// responses. The default behavior, seen in BaseResourceHandler, is to apply no
	// rules.
	Rules() Rules
}

// requestHandler constructs http.HandlerFuncs responsible for handling HTTP requests.
type requestHandler struct {
	API
}

type myReader struct {
	*bytes.Buffer
}

// So that it implements the io.ReadCloser interface
func (m myReader) Close() error { return nil }

// handleCreate returns a HandlerFunc which will deserialize the request payload, pass
// it to the provided create function, and then serialize and dispatch the response.
// The serialization mechanism used is specified by the "format" query parameter.
func (h requestHandler) handleCreate(handler ResourceHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(nil, r)
		version := ctx.Version()
		rules := handler.Rules()

		if r.Header["Content-Type"][0] == "application/x-www-form-urlencoded" {
			var data Payload
			resource, _ := handler.CreateResource(ctx, data, ctx.Version())
			resource = applyOutboundRules(resource, rules, version)
			ctx = ctx.setResult(resource)
			ctx = ctx.setStatus(http.StatusCreated)
		} else {
			data, err := decodePayload(payloadString(r.Body))
			if err != nil {
				ctx = ctx.setError(BadRequest(err.Error()))
			} else {
				data, err := applyInboundRules(data, rules, version)
				if err != nil {
					// Type coercion failed.
					ctx = ctx.setError(UnprocessableRequest(err.Error()))
				} else {
					resource, err := handler.CreateResource(ctx, data, ctx.Version())
					if err == nil {
						resource = applyOutboundRules(resource, rules, version)
					}

					if resource != nil {
						ctx = ctx.setResult(resource)
						ctx = ctx.setStatus(http.StatusCreated)
					} else {
						ctx = ctx.setStatus(http.StatusNoContent)
					}

					if err != nil {
						ctx = ctx.setError(err)
					}
				}
			}
		}
		h.sendResponse(w, ctx)
	})
}

// handleReadList returns a Handler which will pass the request context to the
// provided read function and then serialize and dispatch the response. The
// serialization mechanism used is specified by the "format" query parameter.
func (h requestHandler) handleReadList(handler ResourceHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(nil, r)
		version := ctx.Version()
		rules := handler.Rules()

		resources, cursor, err := handler.ReadResourceList(
			ctx, ctx.Limit(), ctx.Cursor(), version)

		if err == nil {
			// Apply rules to results.
			for idx, resource := range resources {
				resources[idx] = applyOutboundRules(resource, rules, version)
			}
		}

		ctx = ctx.setResult(resources)
		ctx = ctx.setCursor(cursor)
		ctx = ctx.setError(err)
		ctx = ctx.setStatus(http.StatusOK)

		h.sendResponse(w, ctx)
	})
}

// handleRead returns a Handler which will pass the resource id to the provided
// read function and then serialize and dispatch the response. The serialization
// mechanism used is specified by the "format" query parameter.
func (h requestHandler) handleRead(handler ResourceHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(nil, r)
		version := ctx.Version()
		rules := handler.Rules()

		resource, err := handler.ReadResource(ctx, ctx.ResourceID(), version)
		if err == nil {
			resource = applyOutboundRules(resource, rules, version)
		}

		ctx = ctx.setResult(resource)
		ctx = ctx.setError(err)
		ctx = ctx.setStatus(http.StatusOK)

		h.sendResponse(w, ctx)
	})
}

// handleUpdateList returns a Handler which will deserialize the request payload,
// pass it to the provided update function, and then serialize and dispatch the
// response. The serialization mechanism used is specified by the "format" query
// parameter.
func (h requestHandler) handleUpdateList(handler ResourceHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(nil, r)
		version := ctx.Version()
		rules := handler.Rules()

		payloadStr := payloadString(r.Body)
		var data []Payload
		var err error
		data, err = decodePayloadSlice(payloadStr)
		if err != nil {
			var p Payload
			p, err = decodePayload(payloadStr)
			data = []Payload{p}
		}

		if err != nil {
			// Payload decoding failed.
			ctx = ctx.setError(BadRequest(err.Error()))
		} else {
			for i := range data {
				data[i], err = applyInboundRules(data[i], rules, version)
			}
			if err != nil {
				// Type coercion failed.
				ctx = ctx.setError(UnprocessableRequest(err.Error()))
			} else {
				resources, err := handler.UpdateResourceList(ctx, data, version)
				if err == nil {
					// Apply rules to results.
					for idx, resource := range resources {
						resources[idx] = applyOutboundRules(resource, rules, version)
					}
				}

				ctx = ctx.setResult(resources)
				ctx = ctx.setError(err)
				ctx = ctx.setStatus(http.StatusOK)
			}
		}

		h.sendResponse(w, ctx)
	})
}

// handleUpdate returns a Handler which will deserialize the request payload,
// pass it to the provided update function, and then serialize and dispatch the
// response. The serialization mechanism used is specified by the "format" query
// parameter.
func (h requestHandler) handleUpdate(handler ResourceHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(nil, r)
		version := ctx.Version()
		rules := handler.Rules()

		if r.Header["Content-Type"][0] == "application/x-www-form-urlencoded" {
			var data Payload
			resource, _ := handler.CreateResource(ctx, data, ctx.Version())
			resource = applyOutboundRules(resource, rules, version)
			ctx = ctx.setResult(resource)
			ctx = ctx.setStatus(http.StatusCreated)
		} else {
			data, err := decodePayload(payloadString(r.Body))
			if err != nil {
				// Payload decoding failed.
				ctx = ctx.setError(BadRequest(err.Error()))
			} else {
				data, err := applyInboundRules(data, rules, version)
				if err != nil {
					// Type coercion failed.
					ctx = ctx.setError(UnprocessableRequest(err.Error()))
				} else {
					resource, err := handler.UpdateResource(
						ctx, ctx.ResourceID(), data, version)
					if err == nil {
						resource = applyOutboundRules(resource, rules, version)
					}

					ctx = ctx.setResult(resource)
					ctx = ctx.setError(err)
					ctx = ctx.setStatus(http.StatusOK)
				}
			}
		}

		h.sendResponse(w, ctx)
	})
}

// handleDelete returns a Handler which will pass the resource id to the provided
// delete function and then serialize and dispatch the response. The serialization
// mechanism used is specified by the "format" query parameter.
func (h requestHandler) handleDelete(handler ResourceHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(nil, r)
		version := ctx.Version()
		rules := handler.Rules()

		resource, err := handler.DeleteResource(ctx, ctx.ResourceID(), version)
		if err == nil {
			resource = applyOutboundRules(resource, rules, version)
		}

		ctx = ctx.setResult(resource)
		ctx = ctx.setError(err)
		ctx = ctx.setStatus(http.StatusOK)

		h.sendResponse(w, ctx)
	})
}

// sendResponse writes a success or error response to the provided http.ResponseWriter
// based on the contents of the RequestContext.
func (h requestHandler) sendResponse(w http.ResponseWriter, ctx RequestContext) {
	format := ctx.ResponseFormat()
	serializer, err := h.responseSerializer(format)
	if err != nil {
		// Fall back to json serialization.
		serializer = jsonSerializer{}
		ctx = ctx.setError(BadRequest(fmt.Sprintf("Format not implemented: %s", format)))
	}

	sendResponse(w, NewResponse(ctx), serializer)
}

// sendResponse writes a response to the http.ResponseWriter.
func sendResponse(w http.ResponseWriter, r response, serializer ResponseSerializer) {
	status := r.Status
	contentType := serializer.ContentType()

	var response []byte
	if r.Payload != nil {
		var err error
		response, err = serializer.Serialize(r.Payload)
		if err != nil {
			log.Printf("Response serialization failed: %s", err)
			status = http.StatusInternalServerError
			contentType = "text/plain"
			response = []byte(err.Error())
		}
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	w.Write(response)
}

// decodePayload unmarshals the JSON payload and returns the resulting map. If the
// content is empty, an empty map is returned. If decoding fails, nil is returned
// with an error.
func decodePayload(payload []byte) (Payload, error) {
	if len(payload) == 0 {
		return map[string]interface{}{}, nil
	}

	var data Payload
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// decodePayloadSlice unmarshals the JSON payload and returns the resulting slice.
// If the content is empty, an empty list is returned. If decoding fails, nil is
// returned with an error.
func decodePayloadSlice(payload []byte) ([]Payload, error) {
	if len(payload) == 0 {
		return []Payload{}, nil
	}

	var data []Payload
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// payloadString returns the given io.Reader as a string. The reader must be rewound
// after calling this in order to be read again.
func payloadString(payload io.Reader) []byte {
	payloadStr, err := ioutil.ReadAll(payload)
	if err != nil {
		return nil
	}
	return payloadStr
}
