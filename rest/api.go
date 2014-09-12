package rest

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"

	"github.com/gorilla/mux"
)

const (
	defaultLogPrefix     = "rest "
	defaultDocsDirectory = "_docs/"
)

// Address is the address and port to bind to (e.g. ":8080").
type Address string

// FilePath represents a file path.
type FilePath string

// Configuration contains settings for configuring an API.
type Configuration struct {
	Debug         bool
	Logger        *log.Logger
	GenerateDocs  bool
	DocsDirectory string
}

// Debugf prints the formatted string to the Configuration Logger if Debug is enabled.
func (c *Configuration) Debugf(format string, v ...interface{}) {
	if c.Debug {
		c.Logger.Printf(format, v)
	}
}

// NewConfiguration returns a default Configuration.
func NewConfiguration() *Configuration {
	logger := log.New(os.Stdout, defaultLogPrefix, log.LstdFlags)
	return &Configuration{
		Debug:         true,
		Logger:        logger,
		GenerateDocs:  true,
		DocsDirectory: defaultDocsDirectory,
	}
}

// API is the top-level interface encapsulating an HTTP REST server. It's responsible for
// registering ResourceHandlers and routing requests. Use NewAPI to retrieve an instance.
type API interface {
	http.Handler

	// Start begins serving requests. This will block unless it fails, in which case an
	// error will be returned. This will validate any defined Rules. If any Rules are
	// invalid, it will panic.
	Start(Address) error

	// StartTLS begins serving requests received over HTTPS connections. This will block
	// unless it fails, in which case an error will be returned. Files containing a
	// certificate and matching private key for the server must be provided. If the
	// certificate is signed by a certificate authority, the certFile should be the
	// concatenation of the server's certificate followed by the CA's certificate. This
	// will validate any defined Rules. If any Rules are invalid, it will panic.
	StartTLS(Address, FilePath, FilePath) error

	// RegisterResourceHandler binds the provided ResourceHandler to the appropriate REST
	// endpoints and applies any specified middleware. Endpoints will have the following
	// base URL: /api/:version/resourceName.
	RegisterResourceHandler(ResourceHandler, ...RequestMiddleware)

	// RegisterHandlerFunc binds the http.HandlerFunc to the provided URI and applies any
	// specified middleware.
	RegisterHandlerFunc(string, http.HandlerFunc, ...RequestMiddleware)

	// RegisterHandler binds the http.Handler to the provided URI and applies any specified
	// middleware.
	RegisterHandler(string, http.Handler, ...RequestMiddleware)

	// RegisterPathPrefix binds the http.HandlerFunc to URIs matched by the given path
	// prefix and applies any specified middleware.
	RegisterPathPrefix(string, http.HandlerFunc, ...RequestMiddleware)

	// RegisterResponseSerializer registers the provided ResponseSerializer with the given
	// format. If the format has already been registered, it will be overwritten.
	RegisterResponseSerializer(string, ResponseSerializer)

	// UnregisterResponseSerializer unregisters the ResponseSerializer with the provided
	// format. If the format hasn't been registered, this is a no-op.
	UnregisterResponseSerializer(string)

	// AvailableFormats returns a slice containing all of the available serialization
	// formats currently available.
	AvailableFormats() []string

	// Configuration returns the API Configuration.
	Configuration() *Configuration

	// ResourceHandlers returns a slice containing the registered ResourceHandlers.
	ResourceHandlers() []ResourceHandler

	// responseSerializer returns a ResponseSerializer for the given format type. If the
	// format is not implemented, the returned serializer will be nil and the error set.
	responseSerializer(string) (ResponseSerializer, error)
}

// RequestMiddleware is a function that returns a HandlerFunc wrapping the provided HandlerFunc.
// This allows injecting custom logic to operate on requests (e.g. performing authentication).
type RequestMiddleware func(http.HandlerFunc) http.HandlerFunc

// newAuthMiddleware returns a RequestMiddleware used to authenticate requests.
func newAuthMiddleware(authenticate func(*http.Request) error) RequestMiddleware {
	return func(wrapped http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if err := authenticate(r); err != nil {
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
	config             *Configuration
	router             *mux.Router
	mu                 sync.RWMutex
	handler            *requestHandler
	serializerRegistry map[string]ResponseSerializer
	resourceHandlers   []ResourceHandler
}

// NewAPI returns a newly allocated API instance.
func NewAPI(config *Configuration) API {
	r := mux.NewRouter()
	restAPI := &muxAPI{
		config:             config,
		router:             r,
		serializerRegistry: map[string]ResponseSerializer{"json": &jsonSerializer{}},
		resourceHandlers:   make([]ResourceHandler, 0),
	}
	restAPI.handler = &requestHandler{restAPI}
	return restAPI
}

// Start begins serving requests. This will block unless it fails, in which case an error will be
// returned.
func (r *muxAPI) Start(addr Address) error {
	r.preprocess()
	return http.ListenAndServe(string(addr), r.router)
}

// StartTLS begins serving requests received over HTTPS connections. This will block unless it
// fails, in which case an error will be returned. Files containing a certificate and matching
// private key for the server must be provided. If the certificate is signed by a certificate
// authority, the certFile should be the concatenation of the server's certificate followed by
// the CA's certificate.
func (r *muxAPI) StartTLS(addr Address, certFile, keyFile FilePath) error {
	r.preprocess()
	return http.ListenAndServeTLS(string(addr), string(certFile), string(keyFile), r.router)
}

// preprocess performs any necessary preprocessing before the server can be started, including
// Rule validation.
func (r *muxAPI) preprocess() {
	r.validateRules()
	if r.config.GenerateDocs {
		newDocGenerator().generateDocs(r)
	}
}

// RegisterResourceHandler binds the provided ResourceHandler to the appropriate REST endpoints and
// applies any specified middleware. Endpoints will have the following base URL:
// /api/:version/resourceName.
func (r *muxAPI) RegisterResourceHandler(h ResourceHandler, middleware ...RequestMiddleware) {
	h = resourceHandlerProxy{h}
	resource := h.ResourceName()
	middleware = append(middleware, newAuthMiddleware(h.Authenticate))

	r.router.HandleFunc(
		h.CreateURI(), applyMiddleware(r.handler.handleCreate(h), middleware),
	).Methods("POST").Name(resource + ":create")
	r.config.Debugf("Registered create handler at POST %s", h.CreateURI())

	r.router.HandleFunc(
		h.ReadListURI(), applyMiddleware(r.handler.handleReadList(h), middleware),
	).Methods("GET").Name(resource + ":readList")
	r.config.Debugf("Registered read list handler at GET %s", h.ReadListURI())

	r.router.HandleFunc(
		h.ReadURI(), applyMiddleware(r.handler.handleRead(h), middleware),
	).Methods("GET").Name(resource + ":read")
	r.config.Debugf("Registered read handler at GET %s", h.ReadURI())

	r.router.HandleFunc(
		h.UpdateListURI(), applyMiddleware(r.handler.handleUpdateList(h), middleware),
	).Methods("PUT").Name(resource + ":updateList")
	r.config.Debugf("Registered update list handler at PUT %s", h.UpdateListURI())

	r.router.HandleFunc(
		h.UpdateURI(), applyMiddleware(r.handler.handleUpdate(h), middleware),
	).Methods("PUT").Name(resource + ":update")
	r.config.Debugf("Registered update handler at PUT %s", h.UpdateURI())

	r.router.HandleFunc(
		h.DeleteURI(), applyMiddleware(r.handler.handleDelete(h), middleware),
	).Methods("DELETE").Name(resource + ":delete")
	r.config.Debugf("Registered delete handler at DELETE %s", h.DeleteURI())

	r.resourceHandlers = append(r.resourceHandlers, h)
}

// RegisterHandlerFunc binds the http.HandlerFunc to the provided URI and applies any
// specified middleware.
func (r *muxAPI) RegisterHandlerFunc(uri string, handler http.HandlerFunc,
	middleware ...RequestMiddleware) {
	r.router.HandleFunc(uri, applyMiddleware(handler, middleware))
}

// RegisterHandler binds the http.Handler to the provided URI and applies any specified
// middleware.
func (r *muxAPI) RegisterHandler(uri string, handler http.Handler, middleware ...RequestMiddleware) {
	r.router.HandleFunc(uri, applyMiddleware(handler.ServeHTTP, middleware))
}

// RegisterPathPrefix binds the http.HandlerFunc to URIs matched by the given path
// prefix and applies any specified middleware.
func (r *muxAPI) RegisterPathPrefix(uri string, handler http.HandlerFunc,
	middleware ...RequestMiddleware) {
	r.router.PathPrefix(uri).HandlerFunc(applyMiddleware(handler, middleware))
}

// ServeHTTP handles an HTTP request.
func (r *muxAPI) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// RegisterResponseSerializer registers the provided ResponseSerializer with the given format. If the
// format has already been registered, it will be overwritten.
func (r *muxAPI) RegisterResponseSerializer(format string, serializer ResponseSerializer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serializerRegistry[format] = serializer
}

// UnregisterResponseSerializer unregisters the ResponseSerializer with the provided format. If the
// format hasn't been registered, this is a no-op.
func (r *muxAPI) UnregisterResponseSerializer(format string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.serializerRegistry, format)
}

// AvailableFormats returns a slice containing all of the available serialization formats
// currently available.
func (r *muxAPI) AvailableFormats() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	formats := make([]string, 0, len(r.serializerRegistry))
	for format := range r.serializerRegistry {
		formats = append(formats, format)
	}
	sort.Strings(formats)
	return formats
}

// ResourceHandlers returns a slice containing the registered ResourceHandlers.
func (r *muxAPI) ResourceHandlers() []ResourceHandler {
	return r.resourceHandlers
}

// Configuration returns the API Configuration.
func (r *muxAPI) Configuration() *Configuration {
	return r.config
}

// responseSerializer returns a ResponseSerializer for the given format type. If the format
// is not implemented, the returned serializer will be nil and the error set.
func (r *muxAPI) responseSerializer(format string) (ResponseSerializer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if serializer, ok := r.serializerRegistry[format]; ok {
		return serializer, nil
	}
	return nil, fmt.Errorf("Format not implemented: %s", format)
}

// applyMiddleware wraps the HandlerFunc with the provided RequestMiddleware and returns the
// function composition.
func applyMiddleware(h http.HandlerFunc, middleware []RequestMiddleware) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}

// validateRules verifies that the Rules for each ResourceHandler registered with the muxAPI
// are valid, meaning they specify fields that exist and correct types. If a Rule is invalid,
// this will panic.
func (r *muxAPI) validateRules() {
	for _, handler := range r.resourceHandlers {
		rules := handler.Rules()
		if rules == nil || rules.Size() == 0 {
			continue
		}

		if err := rules.Validate(); err != nil {
			panic(err)
		}
	}
}
