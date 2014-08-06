package server

import (
	"encoding/json"
	"fmt"
	"log"
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
	ReadResourceList(context.RequestContext, int, string) ([]Resource, string, error)
	ReadResource(context.RequestContext, string, string) (Resource, error)
	UpdateResource(context.RequestContext, string, Payload, string) (Resource, error)
	DeleteResource(context.RequestContext, string, string) (Resource, error)
	IsAuthorized(http.Request) bool
}

type BaseResourceHandler struct{}

func (b BaseResourceHandler) ResourceName() string {
	panic("ResourceName not implemented")
}

func (b BaseResourceHandler) CreateResource(ctx context.RequestContext, data Payload, version string) (Resource, error) {
	panic("CreateResource not implemented")
}

func (b BaseResourceHandler) ReadResourceList(ctx context.RequestContext, limit int, version string) ([]Resource, string, error) {
	panic("ReadResourceList not implemented")
}

func (b BaseResourceHandler) ReadResource(ctx context.RequestContext, id string, version string) (Resource, error) {
	panic("ReadResource not implemented")
}

func (b BaseResourceHandler) UpdateResource(ctx context.RequestContext, id string, data Payload, version string) (Resource, error) {
	panic("UpdateResource not implemented")
}

func (b BaseResourceHandler) DeleteResource(ctx context.RequestContext, id string, version string) (Resource, error) {
	panic("DeleteResource not implemented")
}

func (b BaseResourceHandler) IsAuthorized(r http.Request) bool {
	return true
}

// RequestMiddleware is a function that returns a HandlerFunc wrapping the provided HandlerFunc.
// This allows injecting custom logic to operate on requests (e.g. performing authentication).
type RequestMiddleware func(http.HandlerFunc) http.HandlerFunc

// newAuthMiddleware returns a RequestMiddleware used to authenticate requests.
func newAuthMiddleware(isAuthorized func(http.Request) bool) RequestMiddleware {
	return func(wrapped http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !isAuthorized(*r) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			wrapped(w, r)
		}
	}
}

// RegisterResourceHandler binds the provided ResourceHandler to the appropriate REST endpoints and
// applies any specified middleware. Endpoints will have the following base URL:
// /api/:version/resourceName.
func RegisterResourceHandler(router *mux.Router, r ResourceHandler, middleware ...RequestMiddleware) {
	urlBase := fmt.Sprintf("/api/v{%s:[^/]+}/%s", context.VersionKey, r.ResourceName())
	resourceUrl := fmt.Sprintf("%s/{%s}", urlBase, context.ResourceIdKey)
	middleware = append(middleware, newAuthMiddleware(r.IsAuthorized))

	router.HandleFunc(
		urlBase,
		applyMiddleware(
			handleCreate(r.CreateResource),
			middleware,
		),
	).Methods("POST").Name("create")
	log.Printf("Registered create handler at POST %s", urlBase)

	router.HandleFunc(
		urlBase,
		applyMiddleware(
			handleReadList(r.ReadResourceList),
			middleware,
		),
	).Methods("GET").Name("readList")
	log.Printf("Registered read list handler at GET %s", urlBase)

	router.HandleFunc(
		resourceUrl,
		applyMiddleware(
			handleRead(r.ReadResource),
			middleware,
		),
	).Methods("GET").Name("read")
	log.Printf("Registered read handler at GET %s", resourceUrl)

	router.HandleFunc(
		resourceUrl,
		applyMiddleware(
			handleUpdate(r.UpdateResource),
			middleware,
		),
	).Methods("PUT").Name("update")
	log.Printf("Registered update handler at UPDATE %s", resourceUrl)

	router.HandleFunc(
		resourceUrl,
		applyMiddleware(
			handleDelete(r.DeleteResource),
			middleware,
		),
	).Methods("DELETE").Name("delete")
	log.Printf("Registered delete handler at DELETE %s", resourceUrl)
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

		SendResponse(w, ctx)
	}
}

// handleReadList returns a HandlerFunc which will pass the request context to the provided read function
// and then serialize and dispatch the response. The serialization mechanism used is specified by the
// "format" query parameter.
func handleReadList(readFunc func(context.RequestContext, int, string) ([]Resource, string, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.NewContext(nil, r)

		resources, cursor, err := readFunc(ctx, ctx.Limit(), ctx.Version())
		ctx = ctx.SetResult(resources)
		ctx = ctx.SetCursor(cursor)
		ctx = ctx.SetError(err)
		ctx = ctx.SetStatus(http.StatusOK)

		SendResponse(w, ctx)
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

		SendResponse(w, ctx)
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

		SendResponse(w, ctx)
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

		SendResponse(w, ctx)
	}
}
