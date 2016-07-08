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
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"

	"github.com/gorilla/mux"
)

type HandleMethod string

const (
	defaultLogPrefix     = "rest "
	defaultDocsDirectory = "_docs/"

	// Handler names
	HandleCreate     HandleMethod = "create"
	HandleRead                    = "read"
	HandleUpdate                  = "update"
	HandleDelete                  = "delete"
	HandleReadList                = "readList"
	HandleUpdateList              = "updateList"
)

// Address is the address and port to bind to (e.g. ":8080").
type Address string

// FilePath represents a file path.
type FilePath string

// An interface satisfied by log.Logger
type StdLogger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})

	Panic(...interface{})
	Panicf(string, ...interface{})
	Panicln(...interface{})
}

// Configuration contains settings for configuring an API.
type Configuration struct {
	Debug         bool
	Logger        StdLogger
	GenerateDocs  bool
	DocsDirectory string
}

// Debugf prints the formatted string to the Configuration Logger if Debug is enabled.
func (c *Configuration) Debugf(format string, v ...interface{}) {
	if c.Debug {
		c.Logger.Printf(format, v...)
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

// MiddlewareError is returned by Middleware to indicate that a request should
// not be served.
type MiddlewareError struct {
	Code     int
	Response []byte
}

// Middleware can be passed in to API#Start and API#StartTLS and will be
// invoked on every request to a route handled by the API. Returns a
// MiddlewareError if the request should be terminated.
type Middleware func(w http.ResponseWriter, r *http.Request) *MiddlewareError

// middlewareProxy proxies an http.Handler by invoking middleware before
// passing the request to the Handler. It implements the http.Handler
// interface.
type middlewareProxy struct {
	handler    http.Handler
	middleware []Middleware
}

// ServeHTTP invokes middleware on the request and then delegates to the
// proxied http.Handler.
func (m *middlewareProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, middleware := range m.middleware {
		if err := middleware(w, r); err != nil {
			w.WriteHeader(err.Code)
			w.Write(err.Response)
			return
		}
	}
	m.handler.ServeHTTP(w, r)
}

// wrapMiddleware returns an http.Handler with the Middleware applied.
func wrapMiddleware(handler http.Handler, middleware ...Middleware) http.Handler {
	return &middlewareProxy{
		handler:    handler,
		middleware: middleware,
	}
}

// API is the top-level interface encapsulating an HTTP REST server. It's responsible for
// registering ResourceHandlers and routing requests. Use NewAPI to retrieve an instance.
type API interface {
	http.Handler

	// Start begins serving requests. This will block unless it fails, in which case an
	// error will be returned. This will validate any defined Rules. If any Rules are
	// invalid, it will panic. Any provided Middleware will be invoked for every request
	// handled by the API.
	Start(Address, ...Middleware) error

	// StartTLS begins serving requests received over HTTPS connections. This will block
	// unless it fails, in which case an error will be returned. Files containing a
	// certificate and matching private key for the server must be provided. If the
	// certificate is signed by a certificate authority, the certFile should be the
	// concatenation of the server's certificate followed by the CA's certificate. This
	// will validate any defined Rules. If any Rules are invalid, it will panic. Any
	// provided Middleware will be invoked for every request handled by the API.
	StartTLS(Address, FilePath, FilePath, ...Middleware) error

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

	// Validate will validate the Rules configured for this API. It returns nil
	// if all Rules are valid, otherwise returns the first encountered
	// validation error.
	Validate() error

	// responseSerializer returns a ResponseSerializer for the given format type. If the
	// format is not implemented, the returned serializer will be nil and the error set.
	responseSerializer(string) (ResponseSerializer, error)
}

// RequestMiddleware is a function that returns a Handler wrapping the provided Handler.
// This allows injecting custom logic to operate on requests (e.g. performing authentication).
type RequestMiddleware func(http.Handler) http.Handler

// newAuthMiddleware returns a RequestMiddleware used to authenticate requests.
func newAuthMiddleware(authenticate func(*http.Request) error) RequestMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := authenticate(r); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(err.Error()))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// newVersionMiddleware checks the request version against all valid versions.
func newVersionMiddleware(validVersions []string) RequestMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestVersion := mux.Vars(r)["version"]

			for _, v := range validVersions {
				if requestVersion == v {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Version %q is not available.", requestVersion)))
		})
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
	restAPI.handler = &requestHandler{restAPI, r}
	return restAPI
}

// Start begins serving requests. This will block unless it fails, in which case an error will be
// returned.
func (r *muxAPI) Start(addr Address, middleware ...Middleware) error {
	r.preprocess()
	return http.ListenAndServe(string(addr), wrapMiddleware(r.router, middleware...))
}

// StartTLS begins serving requests received over HTTPS connections. This will block unless it
// fails, in which case an error will be returned. Files containing a certificate and matching
// private key for the server must be provided. If the certificate is signed by a certificate
// authority, the certFile should be the concatenation of the server's certificate followed by
// the CA's certificate.
func (r *muxAPI) StartTLS(addr Address, certFile, keyFile FilePath, middleware ...Middleware) error {
	r.preprocess()
	return http.ListenAndServeTLS(string(addr), string(certFile), string(keyFile), wrapMiddleware(r.router, middleware...))
}

// preprocess performs any necessary preprocessing before the server can be started, including
// Rule validation.
func (r *muxAPI) preprocess() {
	r.validateRulesOrPanic()
	if r.config.GenerateDocs {
		if err := newDocGenerator().generateDocs(r); err != nil {
			log.Printf("documentation could not be generated: %v", err)
		}
	}
}

// Check the route for an error and log the error if it exists.
func (r *muxAPI) checkRoute(handler, method, uri string, route *mux.Route) {
	err := route.GetError()

	if err != nil {
		log.Printf("Failed to setup route %s with %v", uri, err)
	} else {
		r.config.Debugf("Registered %s handler at %s %s", handler, method, uri)
	}
}

// RegisterResourceHandler binds the provided ResourceHandler to the appropriate REST endpoints and
// applies any specified middleware. Endpoints will have the following base URL:
// /api/:version/resourceName.
func (r *muxAPI) RegisterResourceHandler(h ResourceHandler, middleware ...RequestMiddleware) {
	h = resourceHandlerProxy{h}
	resource := h.ResourceName()
	middleware = append(middleware, newAuthMiddleware(h.Authenticate))
	if validVersions := h.ValidVersions(); validVersions != nil {
		middleware = append(middleware, newVersionMiddleware(validVersions))
	}

	// Some browsers don't support PUT and DELETE, so allow method overriding.
	// POST requests with X-HTTP-Method-Override=PUT/DELETE will route to the
	// respective handlers.

	route := r.router.Handle(
		h.ReadListURI(), applyMiddleware(r.handler.handleReadList(h), middleware),
	).Methods("POST").Headers("X-HTTP-Method-Override", "GET").Name(resource + ":readListOverride")
	r.checkRoute("read list override", h.ReadListURI(), "OVERRIDE-GET", route)

	route = r.router.Handle(
		h.ReadURI(), applyMiddleware(r.handler.handleRead(h), middleware),
	).Methods("POST").Headers("X-HTTP-Method-Override", "GET").Name(resource + ":readOverride")
	r.checkRoute("read override", h.ReadURI(), "OVERRIDE-GET", route)

	route = r.router.Handle(
		h.UpdateListURI(), applyMiddleware(r.handler.handleUpdateList(h), middleware),
	).Methods("POST").Headers("X-HTTP-Method-Override", "PUT").Name(resource + ":updateListOverride")
	r.checkRoute("update list override", h.UpdateListURI(), "OVERRIDE-PUT", route)

	route = r.router.Handle(
		h.UpdateURI(), applyMiddleware(r.handler.handleUpdate(h), middleware),
	).Methods("POST").Headers("X-HTTP-Method-Override", "PUT").Name(resource + ":updateOverride")
	r.checkRoute("update override", h.UpdateURI(), "OVERRIDE-PUT", route)

	route = r.router.Handle(
		h.DeleteURI(), applyMiddleware(r.handler.handleDelete(h), middleware),
	).Methods("POST").Headers("X-HTTP-Method-Override", "DELETE").Name(resource + ":deleteOverride")
	r.checkRoute("delete override", h.DeleteURI(), "OVERRIDE-DELETE", route)

	// These return a Route which has a GetError command. Probably should check
	// that and log it if it fails :)
	r.router.Handle(
		h.CreateURI(), applyMiddleware(r.handler.handleCreate(h), middleware),
	).Methods("POST").Name(resource + ":" + string(HandleCreate))
	r.checkRoute("create", h.CreateURI(), "POST", route)

	r.router.Handle(
		h.ReadListURI(), applyMiddleware(r.handler.handleReadList(h), middleware),
	).Methods("GET").Name(resource + ":" + string(HandleReadList))
	r.checkRoute("read list", h.ReadListURI(), "GET", route)

	r.router.Handle(
		h.ReadURI(), applyMiddleware(r.handler.handleRead(h), middleware),
	).Methods("GET").Name(resource + ":" + string(HandleRead))
	r.checkRoute("read", h.ReadURI(), "GET", route)

	r.router.Handle(
		h.UpdateListURI(), applyMiddleware(r.handler.handleUpdateList(h), middleware),
	).Methods("PUT").Name(resource + ":" + string(HandleUpdateList))
	r.checkRoute("update list", h.UpdateListURI(), "PUT", route)

	r.router.Handle(
		h.UpdateURI(), applyMiddleware(r.handler.handleUpdate(h), middleware),
	).Methods("PUT").Name(resource + ":" + string(HandleUpdate))
	r.checkRoute("update", h.UpdateURI(), "PUT", route)

	r.router.Handle(
		h.DeleteURI(), applyMiddleware(r.handler.handleDelete(h), middleware),
	).Methods("DELETE").Name(resource + ":" + string(HandleDelete))
	r.checkRoute("delete", h.DeleteURI(), "DELETE", route)

	r.resourceHandlers = append(r.resourceHandlers, h)
}

// RegisterHandlerFunc binds the http.HandlerFunc to the provided URI and applies any
// specified middleware.
func (r *muxAPI) RegisterHandlerFunc(uri string, handlerfunc http.HandlerFunc,
	middleware ...RequestMiddleware) {
	r.router.Handle(uri, applyMiddleware(http.HandlerFunc(handlerfunc), middleware))
}

// RegisterHandler binds the http.Handler to the provided URI and applies any specified
// middleware.
func (r *muxAPI) RegisterHandler(uri string, handler http.Handler, middleware ...RequestMiddleware) {
	r.router.Handle(uri, applyMiddleware(handler, middleware))
}

// RegisterPathPrefix binds the http.HandlerFunc to URIs matched by the given path
// prefix and applies any specified middleware.
func (r *muxAPI) RegisterPathPrefix(uri string, handler http.HandlerFunc,
	middleware ...RequestMiddleware) {
	r.router.PathPrefix(uri).Handler(applyMiddleware(handler, middleware))
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

// Validate will validate the Rules configured for this API. It returns nil if
// all Rules are valid, otherwise returns the first encountered validation
// error.
func (r *muxAPI) Validate() error {
	for _, handler := range r.resourceHandlers {
		rules := handler.Rules()
		if rules == nil || rules.Size() == 0 {
			continue
		}

		if err := rules.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// validateRulesOrPanic verifies that the Rules for each ResourceHandler
// registered with the muxAPI are valid, meaning they specify fields that exist
// and correct types. If a Rule is invalid, this will panic.
func (r *muxAPI) validateRulesOrPanic() {
	if err := r.Validate(); err != nil {
		panic(err)
	}
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

// applyMiddleware wraps the Handler with the provided RequestMiddleware and returns another Handler.
func applyMiddleware(h http.Handler, middleware []RequestMiddleware) http.Handler {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}
