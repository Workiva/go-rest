package rest

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/gorilla/mux"
)

// API is the top-level interface encapsulating an HTTP REST server. It's responsible for
// registering ResourceHandlers and routing requests. Use NewAPI to retrieve an instance.
type API interface {
	// Start begins serving requests. This will block unless it fails, in which case an
	// error will be returned.
	Start(string) error

	// StartTLS begins serving requests received over HTTPS connections. This will block
	// unless it fails, in which case an error will be returned. Files containing a
	// certificate and matching private key for the server must be provided. If the
	// certificate is signed by a certificate authority, the certFile should be the
	// concatenation of the server's certificate followed by the CA's certificate.
	StartTLS(string, string, string) error

	// RegisterResourceHandler binds the provided ResourceHandler to the appropriate REST
	// endpoints and applies any specified middleware. Endpoints will have the following
	// base URL: /api/:version/resourceName.
	RegisterResourceHandler(ResourceHandler, ...RequestMiddleware)

	// RegisterResponseSerializer registers the provided ResponseSerializer with the given
	// format. If the format has already been registered, it will be overwritten.
	RegisterResponseSerializer(string, ResponseSerializer)

	// UnregisterResponseSerializer unregisters the ResponseSerializer with the provided
	// format. If the format hasn't been registered, this is a no-op.
	UnregisterResponseSerializer(string)

	// AvailableFormats returns a slice containing all of the available serialization
	// formats currently available.
	AvailableFormats() []string

	// responseSerializer returns a ResponseSerializer for the given format type. If the
	// format is not implemented, the returned serializer will be nil and the error set.
	responseSerializer(string) (ResponseSerializer, error)
}

// RequestMiddleware is a function that returns a HandlerFunc wrapping the provided HandlerFunc.
// This allows injecting custom logic to operate on requests (e.g. performing authentication).
type RequestMiddleware func(http.HandlerFunc) http.HandlerFunc

// newAuthMiddleware returns a RequestMiddleware used to authenticate requests.
func newAuthMiddleware(authenticate func(http.Request) error) RequestMiddleware {
	return func(wrapped http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if err := authenticate(*r); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(err.Error()))
				return
			}
			wrapped(w, r)
		}
	}
}

// muxAPI is an implementation of the API interface which relies on the gorilla/mux
// package to handle request dispatching (see http://www.gorillatoolkit.org/pkg/mux).
type muxAPI struct {
	router             *mux.Router
	mu                 sync.RWMutex
	handler            *requestHandler
	serializerRegistry map[string]ResponseSerializer
}

// NewAPI returns a newly allocated API instance.
func NewAPI() API {
	r := mux.NewRouter()
	restAPI := &muxAPI{
		router:             r,
		serializerRegistry: map[string]ResponseSerializer{"json": &jsonSerializer{}},
	}
	restAPI.handler = &requestHandler{restAPI}
	return restAPI
}

// Start begins serving requests. This will block unless it fails, in which case an error will be
// returned.
func (r muxAPI) Start(addr string) error {
	return http.ListenAndServe(addr, r.router)
}

// StartTLS begins serving requests received over HTTPS connections. This will block unless it
// fails, in which case an error will be returned. Files containing a certificate and matching
// private key for the server must be provided. If the certificate is signed by a certificate
// authority, the certFile should be the concatenation of the server's certificate followed by
// the CA's certificate.
func (r muxAPI) StartTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, r.router)
}

// RegisterResourceHandler binds the provided ResourceHandler to the appropriate REST endpoints and
// applies any specified middleware. Endpoints will have the following base URL:
// /api/:version/resourceName.
func (r muxAPI) RegisterResourceHandler(h ResourceHandler, middleware ...RequestMiddleware) {
	resource := h.ResourceName()
	urlBase := fmt.Sprintf("/api/v{%s:[^/]+}/%s", versionKey, resource)
	resourceURL := fmt.Sprintf("%s/{%s}", urlBase, resourceIDKey)
	middleware = append(middleware, newAuthMiddleware(h.Authenticate))

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
		resourceURL,
		applyMiddleware(
			r.handler.handleRead(h.ReadResource),
			middleware,
		),
	).Methods("GET").Name(resource + ":read")
	log.Printf("Registered read handler at GET %s", resourceURL)

	r.router.HandleFunc(
		resourceURL,
		applyMiddleware(
			r.handler.handleUpdate(h.UpdateResource),
			middleware,
		),
	).Methods("PUT").Name(resource + ":update")
	log.Printf("Registered update handler at UPDATE %s", resourceURL)

	r.router.HandleFunc(
		resourceURL,
		applyMiddleware(
			r.handler.handleDelete(h.DeleteResource),
			middleware,
		),
	).Methods("DELETE").Name(resource + ":delete")
	log.Printf("Registered delete handler at DELETE %s", resourceURL)
}

// RegisterResponseSerializer registers the provided ResponseSerializer with the given format. If the
// format has already been registered, it will be overwritten.
func (r muxAPI) RegisterResponseSerializer(format string, serializer ResponseSerializer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serializerRegistry[format] = serializer
}

// UnregisterResponseSerializer unregisters the ResponseSerializer with the provided format. If the
// format hasn't been registered, this is a no-op.
func (r muxAPI) UnregisterResponseSerializer(format string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.serializerRegistry, format)
}

// AvailableFormats returns a slice containing all of the available serialization formats
// currently available.
func (r muxAPI) AvailableFormats() []string {
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
func (r muxAPI) responseSerializer(format string) (ResponseSerializer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if serializer, ok := r.serializerRegistry[format]; ok {
		return serializer, nil
	}
	return nil, fmt.Errorf("Format not implemented: %s", format)
}

// getRouteHandler returns the http.Handler for the API route with the given name.
// This is purely for testing purposes and shouldn't be used elsewhere.
func (r muxAPI) getRouteHandler(name string) (http.Handler, error) {
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
