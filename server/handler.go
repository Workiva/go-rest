package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go-rest/server/context"

	"github.com/gorilla/mux"
)

// Resource represents a domain model.
type Resource interface{}

// Payload is the unmarshalled request body.
type Payload map[string]interface{}

// ResourceHandler specifies the endpoint handlers for working with a resource. This consists of
// the business logic for performing CRUD operations.
type ResourceHandler interface {
	ResourceName() string
	CreateResource(context.RequestContext, Payload, string) (Resource, error)
	ReadResourceList(context.RequestContext, string) ([]Resource, string, error)
	ReadResource(context.RequestContext, string, string) (Resource, error)
	UpdateResource(context.RequestContext, string, Payload, string) (Resource, error)
	DeleteResource(context.RequestContext, string, string) (Resource, error)
}

// RequestMiddleware is a function that returns a HandlerFunc wrapping the provided HandlerFunc.
// This allows injecting custom logic to operate on requests (e.g. performing authentication).
type RequestMiddleware func(http.HandlerFunc) http.HandlerFunc

// RegisterResourceHandler binds the provided ResourceHandler to the appropriate REST endpoints and
// applies any specified middleware. Endpoints will have the following base URL:
// /api/:version/resourceName.
func RegisterResourceHandler(router *mux.Router, r ResourceHandler, middleware ...RequestMiddleware) {
	urlBase := fmt.Sprintf("/api/v{%s:[^/]+}/%s", context.VersionKey, r.ResourceName())
	resourceUrl := fmt.Sprintf("%s/{%s}", urlBase, context.ResourceIdKey)

	router.HandleFunc(
		urlBase,
		applyMiddleware(
			handleCreate(r.CreateResource),
			middleware,
		),
	).Methods("POST").Name("create")

	router.HandleFunc(
		urlBase,
		applyMiddleware(
			handleReadList(r.ReadResourceList),
			middleware,
		),
	).Methods("GET").Name("readList")

	router.HandleFunc(
		resourceUrl,
		applyMiddleware(
			handleRead(r.ReadResource),
			middleware,
		),
	).Methods("GET").Name("read")

	router.HandleFunc(
		resourceUrl,
		applyMiddleware(
			handleUpdate(r.UpdateResource),
			middleware,
		),
	).Methods("PUT").Name("update")

	router.HandleFunc(
		resourceUrl,
		applyMiddleware(
			handleDelete(r.DeleteResource),
			middleware,
		),
	).Methods("DELETE").Name("delete")
}

// applyMiddleware wraps the HandlerFunc with the provided RequestMiddleware and returns the
// function composition.
func applyMiddleware(h http.HandlerFunc, middleware []RequestMiddleware) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}

// handleCreate returns a HandlerFunc which will deserialize the request payload, pass it to the
// provided create function, and then serialize and dispatch the response. The
// serialization mechanism used is specified by the "format" query parameter.
func handleCreate(createFunc func(context.RequestContext, Payload, string) (Resource, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.NewContext(nil, r)

		decoder := json.NewDecoder(r.Body)
		var data map[string]interface{}
		if err := decoder.Decode(&data); err != nil {
			ctx = ctx.SetError(err)
			ctx = ctx.SetStatus(http.StatusInternalServerError)
		} else {
			resource, err := createFunc(ctx, data, ctx.Version())
			ctx = ctx.SetResult(resource)
			ctx = ctx.SetStatus(http.StatusCreated)
			if err != nil {
				ctx = ctx.SetError(err)
			}
		}

		sendResponse(w, ctx)
	}
}

// handleReadList returns a HandlerFunc which will pass the request context to the provided read function
// and then serialize and dispatch the response. The serialization mechanism used is specified by the
// "format" query parameter.
func handleReadList(readFunc func(context.RequestContext, string) ([]Resource, string, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.NewContext(nil, r)

		resources, cursor, err := readFunc(ctx, ctx.Version())
		ctx = ctx.SetResult(resources)
		ctx = ctx.SetCursor(cursor)
		ctx = ctx.SetError(err)
		ctx = ctx.SetStatus(http.StatusOK)

		sendResponse(w, ctx)
	}
}

// handleRead returns a HandlerFunc which will pass the resource id to the provided read function
// and then serialize and dispatch the response. The serialization mechanism used is specified by
// the "format" query parameter.
func handleRead(readFunc func(context.RequestContext, string, string) (Resource, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.NewContext(nil, r)

		resource, err := readFunc(ctx, ctx.ResourceId(), ctx.Version())
		ctx = ctx.SetResult(resource)
		ctx = ctx.SetError(err)
		ctx = ctx.SetStatus(http.StatusOK)

		sendResponse(w, ctx)
	}
}

// handleUpdate returns a HandlerFunc which will deserialize the request payload, pass it to the
// provided update function, and then serialize and dispatch the response. The serialization
// mechanism used is specified by the "format" query parameter.
func handleUpdate(updateFunc func(context.RequestContext, string, Payload, string) (Resource, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.NewContext(nil, r)

		decoder := json.NewDecoder(r.Body)
		var data map[string]interface{}
		if err := decoder.Decode(&data); err != nil {
			ctx = ctx.SetError(err)
			ctx = ctx.SetStatus(http.StatusInternalServerError)
		} else {
			resource, err := updateFunc(ctx, ctx.ResourceId(), data, ctx.Version())
			ctx = ctx.SetResult(resource)
			ctx = ctx.SetError(err)
			ctx = ctx.SetStatus(http.StatusOK)
		}

		sendResponse(w, ctx)
	}
}

// handleDelete returns a HandlerFunc which will pass the resource id to the provided delete
// function and then serialize and dispatch the response. The serialization mechanism used
// is specified by the "format" query parameter.
func handleDelete(deleteFunc func(context.RequestContext, string, string) (Resource, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.NewContext(nil, r)

		resource, err := deleteFunc(ctx, ctx.ResourceId(), ctx.Version())
		ctx = ctx.SetResult(resource)
		ctx = ctx.SetError(err)
		ctx = ctx.SetStatus(http.StatusOK)

		sendResponse(w, ctx)
	}
}
