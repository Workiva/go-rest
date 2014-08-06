package server

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"

	"go-rest/server/context"

	"github.com/gorilla/mux"
)

// RestApi is the top-level interface encapsulating an HTTP REST server. It's responsible for
// registering ResourceHandlers and routing requests. Use NewRestApi to retrieve an instance.
type RestApi interface {
	Start(addr string)
	RegisterResourceHandler(ResourceHandler, ...RequestMiddleware)
	AddResponseSerializer(string, ResponseSerializer)
	RemoveResponseSerializer(string)
	AvailableFormats() []string
	responseSerializer(string) (ResponseSerializer, error)
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

// muxRestApi is an implementation of the RestApi interface which relies on the gorilla/mux
// package to handle request dispatching (see http://www.gorillatoolkit.org/pkg/mux).
type muxRestApi struct {
	router             *mux.Router
	mu                 sync.RWMutex
	handler            *requestHandler
	serializerRegistry map[string]ResponseSerializer
}

// NewRestApi returns a newly allocated RestApi instance.
func NewRestApi() RestApi {
	r := mux.NewRouter()
	restApi := &muxRestApi{
		router:             r,
		serializerRegistry: map[string]ResponseSerializer{"json": &jsonSerializer{}},
	}
	restApi.handler = &requestHandler{restApi}
	return restApi
}

// Start begins serving requests. This will block.
func (r muxRestApi) Start(addr string) {
	http.ListenAndServe(addr, r.router)
}

// RegisterResourceHandler binds the provided ResourceHandler to the appropriate REST endpoints and
// applies any specified middleware. Endpoints will have the following base URL:
// /api/:version/resourceName.
func (r muxRestApi) RegisterResourceHandler(h ResourceHandler, middleware ...RequestMiddleware) {
	resource := h.ResourceName()
	urlBase := fmt.Sprintf("/api/v{%s:[^/]+}/%s", context.VersionKey, resource)
	resourceUrl := fmt.Sprintf("%s/{%s}", urlBase, context.ResourceIdKey)
	middleware = append(middleware, newAuthMiddleware(h.IsAuthorized))

	r.router.HandleFunc(
		urlBase,
		applyMiddleware(
			r.handler.handleCreate(h.CreateResource),
			middleware,
		),
	).Methods("POST").Name(resource + ":create")
	log.Printf("Registered create handler at POST %s", urlBase)

	r.router.HandleFunc(
		urlBase,
		applyMiddleware(
			r.handler.handleReadList(h.ReadResourceList),
			middleware,
		),
	).Methods("GET").Name(resource + ":readList")
	log.Printf("Registered read list handler at GET %s", urlBase)

	r.router.HandleFunc(
		resourceUrl,
		applyMiddleware(
			r.handler.handleRead(h.ReadResource),
			middleware,
		),
	).Methods("GET").Name(resource + ":read")
	log.Printf("Registered read handler at GET %s", resourceUrl)

	r.router.HandleFunc(
		resourceUrl,
		applyMiddleware(
			r.handler.handleUpdate(h.UpdateResource),
			middleware,
		),
	).Methods("PUT").Name(resource + ":update")
	log.Printf("Registered update handler at UPDATE %s", resourceUrl)

	r.router.HandleFunc(
		resourceUrl,
		applyMiddleware(
			r.handler.handleDelete(h.DeleteResource),
			middleware,
		),
	).Methods("DELETE").Name(resource + ":delete")
	log.Printf("Registered delete handler at DELETE %s", resourceUrl)
}

// AddResponseSerializer registers the provided ResponseSerializer with the given format. If the
// format has already been registered, it will be overwritten.
func (r muxRestApi) AddResponseSerializer(format string, serializer ResponseSerializer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serializerRegistry[format] = serializer
}

// RemoveResponseSerializer unregisters the ResponseSerializer with the provided format. If the
// format hasn't been registered, this is a no-op.
func (r muxRestApi) RemoveResponseSerializer(format string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.serializerRegistry, format)
}

// AvailableFormats returns a slice containing all of the available serialization formats
// currently available.
func (r muxRestApi) AvailableFormats() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	formats := make([]string, 0, len(r.serializerRegistry))
	for format := range r.serializerRegistry {
		formats = append(formats, format)
	}
	sort.Strings(formats)
	return formats
}

// responseSerializer returns a ResponseSerializer for the given format type. If the format
// is not implemented, the returned serializer will be nil and the error set.
func (r muxRestApi) responseSerializer(format string) (ResponseSerializer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if serializer, ok := r.serializerRegistry[format]; ok {
		return serializer, nil
	}
	return nil, fmt.Errorf("Format not implemented: %s", format)
}

// getRouteHandler returns the http.Handler for the API route with the given name.
// This is purely for testing purposes and shouldn't be used elsewhere.
func (r muxRestApi) getRouteHandler(name string) (http.Handler, error) {
	route := r.router.Get(name)
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
