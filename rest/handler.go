package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Resource represents a domain model.
type Resource interface{}

// Payload is the unmarshalled request body.
type Payload map[string]interface{}

// ResourceHandler specifies the endpoint handlers for working with a resource. This
// consists of the business logic for performing CRUD operations.
type ResourceHandler interface {
	// ResourceName is used to identify what resource a handler corresponds to and is
	// used in the endpoint URLs, i.e. /api/:version/resourceName.
	ResourceName() string

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
	Authenticate(http.Request) error

	// Rules returns the resource rules to apply to incoming requests and outgoing
	// responses. The default behavior, seen in BaseResourceHandler, is to apply no
	// rules. Different Rules can be returned based on the version provided.
	Rules(string) []Rule
}

// BaseResourceHandler is a base implementation of ResourceHandler with stubs for the
// CRUD operations. This allows ResourceHandler implementations to only implement
// what they need.
type BaseResourceHandler struct{}

// ResourceName is a stub. It must be implemented.
func (b BaseResourceHandler) ResourceName() string {
	panic("ResourceName not implemented")
}

// CreateResource is a stub. Implement if necessary.
func (b BaseResourceHandler) CreateResource(ctx RequestContext, data Payload,
	version string) (Resource, error) {
	panic("CreateResource not implemented")
}

// ReadResourceList is a stub. Implement if necessary.
func (b BaseResourceHandler) ReadResourceList(ctx RequestContext, limit int,
	cursor string, version string) ([]Resource, string, error) {
	panic("ReadResourceList not implemented")
}

// ReadResource is a stub. Implement if necessary.
func (b BaseResourceHandler) ReadResource(ctx RequestContext, id string,
	version string) (Resource, error) {
	panic("ReadResource not implemented")
}

// UpdateResource is a stub. Implement if necessary.
func (b BaseResourceHandler) UpdateResource(ctx RequestContext, id string,
	data Payload, version string) (Resource, error) {
	panic("UpdateResource not implemented")
}

// DeleteResource is a stub. Implement if necessary.
func (b BaseResourceHandler) DeleteResource(ctx RequestContext, id string,
	version string) (Resource, error) {
	panic("DeleteResource not implemented")
}

// Authenticate is the default authentication logic. All requests are authorized.
// Implement custom authentication logic if necessary.
func (b BaseResourceHandler) Authenticate(r http.Request) error {
	return nil
}

// Rules returns the resource rules to apply to incoming requests and outgoing
// responses. No rules are applied by default. Implement if necessary.
func (b BaseResourceHandler) Rules(version string) []Rule {
	return []Rule{}
}

// requestHandler constructs http.HandlerFuncs responsible for handling HTTP requests.
type requestHandler struct {
	API
}

// handleCreate returns a HandlerFunc which will deserialize the request payload, pass
// it to the provided create function, and then serialize and dispatch the response.
// The serialization mechanism used is specified by the "format" query parameter.
func (h requestHandler) handleCreate(handler ResourceHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(nil, r)
		version := ctx.Version()

		decoder := json.NewDecoder(r.Body)
		var data map[string]interface{}
		if err := decoder.Decode(&data); err != nil {
			// Payload decoding failed.
			ctx = ctx.setError(err)
			ctx = ctx.setStatus(http.StatusInternalServerError)
		} else {
			data, err := applyInboundRules(data, handler.Rules(version))
			if err != nil {
				// Type coercion failed.
				ctx = ctx.setError(UnprocessableRequest(err.Error()))
			} else {
				resource, err := handler.CreateResource(ctx, data, ctx.Version())
				resource = applyOutboundRules(resource, handler.Rules(ctx.Version()))
				ctx = ctx.setResult(resource)
				ctx = ctx.setStatus(http.StatusCreated)
				if err != nil {
					ctx = ctx.setError(err)
				}
			}
		}

		h.sendResponse(w, ctx)
	}
}

// handleReadList returns a HandlerFunc which will pass the request context to the
// provided read function and then serialize and dispatch the response. The
// serialization mechanism used is specified by the "format" query parameter.
func (h requestHandler) handleReadList(handler ResourceHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(nil, r)
		version := ctx.Version()

		resources, cursor, err := handler.ReadResourceList(
			ctx, ctx.Limit(), ctx.Cursor(), version)

		// Apply rules to results.
		for idx, resource := range resources {
			resources[idx] = applyOutboundRules(resource, handler.Rules(version))
		}

		ctx = ctx.setResult(resources)
		ctx = ctx.setCursor(cursor)
		ctx = ctx.setError(err)
		ctx = ctx.setStatus(http.StatusOK)

		h.sendResponse(w, ctx)
	}
}

// handleRead returns a HandlerFunc which will pass the resource id to the provided
// read function and then serialize and dispatch the response. The serialization
// mechanism used is specified by the "format" query parameter.
func (h requestHandler) handleRead(handler ResourceHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(nil, r)
		version := ctx.Version()

		resource, err := handler.ReadResource(ctx, ctx.ResourceID(), version)
		rules := handler.Rules(version)
		resource = applyOutboundRules(resource, rules)

		ctx = ctx.setResult(resource)
		ctx = ctx.setError(err)
		ctx = ctx.setStatus(http.StatusOK)

		h.sendResponse(w, ctx)
	}
}

// handleUpdate returns a HandlerFunc which will deserialize the request payload,
// pass it to the provided update function, and then serialize and dispatch the
// response. The serialization mechanism used is specified by the "format" query
// parameter.
func (h requestHandler) handleUpdate(handler ResourceHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(nil, r)
		version := ctx.Version()

		decoder := json.NewDecoder(r.Body)
		var data map[string]interface{}
		if err := decoder.Decode(&data); err != nil {
			// Payload decoding failed.
			ctx = ctx.setError(err)
			ctx = ctx.setStatus(http.StatusInternalServerError)
		} else {
			data, err := applyInboundRules(data, handler.Rules(version))
			if err != nil {
				// Type coercion failed.
				ctx = ctx.setError(UnprocessableRequest(err.Error()))
			} else {
				resource, err := handler.UpdateResource(
					ctx, ctx.ResourceID(), data, version)
				rules := handler.Rules(version)
				resource = applyOutboundRules(resource, rules)

				ctx = ctx.setResult(resource)
				ctx = ctx.setError(err)
				ctx = ctx.setStatus(http.StatusOK)
			}
		}

		h.sendResponse(w, ctx)
	}
}

// handleDelete returns a HandlerFunc which will pass the resource id to the provided
// delete function and then serialize and dispatch the response. The serialization
// mechanism used is specified by the "format" query parameter.
func (h requestHandler) handleDelete(handler ResourceHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(nil, r)
		version := ctx.Version()

		resource, err := handler.DeleteResource(ctx, ctx.ResourceID(), version)
		rules := handler.Rules(version)
		resource = applyOutboundRules(resource, rules)

		ctx = ctx.setResult(resource)
		ctx = ctx.setError(err)
		ctx = ctx.setStatus(http.StatusOK)

		h.sendResponse(w, ctx)
	}
}

// sendResponse writes a success or error response to the provided http.ResponseWriter
// based on the contents of the RequestContext.
func (h requestHandler) sendResponse(w http.ResponseWriter, ctx RequestContext) {
	status := ctx.Status()
	requestError := ctx.Error()
	result := ctx.Result()
	format := ctx.ResponseFormat()

	serializer, err := h.responseSerializer(format)
	if err != nil {
		// Fall back to json serialization.
		serializer = jsonSerializer{}
		requestError = NotImplemented(fmt.Sprintf("Format not implemented: %s", format))
	}

	var response response
	if requestError != nil {
		response = NewErrorResponse(requestError)
	} else {
		nextURL, _ := ctx.NextURL()
		response = NewSuccessResponse(result, status, nextURL)
	}

	sendResponse(w, response, serializer)
}

// sendResponse writes a response to the http.ResponseWriter.
func sendResponse(w http.ResponseWriter, r response, serializer ResponseSerializer) {
	status := r.Status
	contentType := serializer.ContentType()
	response, err := serializer.Serialize(r.Payload)
	if err != nil {
		log.Printf("Response serialization failed: %s", err)
		status = http.StatusInternalServerError
		contentType = "text/plain"
		response = []byte(err.Error())
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	w.Write(response)
}
