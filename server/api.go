package server

import (
	"fmt"
	"log"
	"net/http"

	"go-rest/server/context"

	"github.com/gorilla/mux"
)

// RestApi is the top-level interface encapsulating an HTTP REST server. It's responsible for
// registering ResourceHandlers and routing requests. Use NewRestApi to retrieve an instance.
type RestApi interface {
	Start(addr string)
	RegisterResourceHandler(ResourceHandler, ...RequestMiddleware)
}

// muxRestApi is an implementation of the RestApi interface which relies on the gorilla/mux
// package to handle request dispatching (see http://www.gorillatoolkit.org/pkg/mux).
type muxRestApi struct {
	*mux.Router
}

// NewRestApi returns a newly allocated RestApi instance.
func NewRestApi() RestApi {
	r := mux.NewRouter()
	return &muxRestApi{r}
}

// Start begins serving requests. This will block.
func (r muxRestApi) Start(addr string) {
	http.ListenAndServe(addr, r)
}

// RegisterResourceHandler binds the provided ResourceHandler to the appropriate REST endpoints and
// applies any specified middleware. Endpoints will have the following base URL:
// /api/:version/resourceName.
func (r muxRestApi) RegisterResourceHandler(h ResourceHandler, middleware ...RequestMiddleware) {
	resource := h.ResourceName()
	urlBase := fmt.Sprintf("/api/v{%s:[^/]+}/%s", context.VersionKey, resource)
	resourceUrl := fmt.Sprintf("%s/{%s}", urlBase, context.ResourceIdKey)
	middleware = append(middleware, newAuthMiddleware(h.IsAuthorized))

	r.HandleFunc(
		urlBase,
		applyMiddleware(
			handleCreate(h.CreateResource),
			middleware,
		),
	).Methods("POST").Name(resource + ":create")
	log.Printf("Registered create handler at POST %s", urlBase)

	r.HandleFunc(
		urlBase,
		applyMiddleware(
			handleReadList(h.ReadResourceList),
			middleware,
		),
	).Methods("GET").Name(resource + ":readList")
	log.Printf("Registered read list handler at GET %s", urlBase)

	r.HandleFunc(
		resourceUrl,
		applyMiddleware(
			handleRead(h.ReadResource),
			middleware,
		),
	).Methods("GET").Name(resource + ":read")
	log.Printf("Registered read handler at GET %s", resourceUrl)

	r.HandleFunc(
		resourceUrl,
		applyMiddleware(
			handleUpdate(h.UpdateResource),
			middleware,
		),
	).Methods("PUT").Name(resource + ":update")
	log.Printf("Registered update handler at UPDATE %s", resourceUrl)

	r.HandleFunc(
		resourceUrl,
		applyMiddleware(
			handleDelete(h.DeleteResource),
			middleware,
		),
	).Methods("DELETE").Name(resource + ":delete")
	log.Printf("Registered delete handler at DELETE %s", resourceUrl)
}

// GetRouteHandler returns the http.Handler for the API route with the given name.
// This is purely for testing purposes and shouldn't be used.
func (r muxRestApi) GetRouteHandler(name string) (http.Handler, error) {
	route := r.Get(name)
	if route == nil {
		return nil, fmt.Errorf("No API route with name %s", name)
	}

	return route.GetHandler(), nil
}

// applyMiddleware wraps the HandlerFunc with the provided RequestMiddleware and returns the
// function composition.
func applyMiddleware(h http.HandlerFunc, middleware []RequestMiddleware) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}
