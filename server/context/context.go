package context

import (
	"net/http"

	"code.google.com/p/go.net/context"
	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
)

const (
	RequestKey    = 0
	ResourceIdKey = "resource_id"
	FormatKey     = "format"
	VersionKey    = "version"
)

// RequestContext contains the context information for the current HTTP request. It's a wrapper
// around Google's Context (http://godoc.org/code.google.com/p/go.net/context), which provides
// facilities for sending request-scoped values, cancelation signals, and deadlines
// across API boundaries to all the goroutines involved in handling a request.
type RequestContext interface {
	context.Context
	ValueWithDefault(key, defaultVal interface{}) interface{}
	ResponseFormat() string
	ResourceId() string
	Version() string
}

// requestContext is an implementation of the RequestContext interface.
type requestContext struct {
	context.Context
	req *http.Request
}

// NewContext returns a RequestContext populated with parameters from the request path and
// query string.
func NewContext(parent context.Context, req *http.Request) RequestContext {
	if parent == nil {
		parent = context.Background()
	}

	for key, value := range req.URL.Query() {
		var val interface{}
		val = value

		// Query string values are slices (e.g. ?foo=bar,baz,qux yields
		// [bar, baz, qux] for foo), but we unbox single values (e.g. ?foo=bar
		// yields bar for foo).
		if len(value) == 1 {
			val = value[0]
		}

		gcontext.Set(req, key, val)
	}

	for key, value := range mux.Vars(req) {
		gcontext.Set(req, key, value)
	}

	// TODO: Keys can potentially be overwritten if the request path has
	// parameters with the same name as query string values. Figure out a
	// better way to handle this.

	return &requestContext{parent, req}
}

// Value returns Gorilla's context package's value for this Context's request
// and key. It delegates to the parent Context if there is no such value.
func (ctx *requestContext) Value(key interface{}) interface{} {
	if key == RequestKey {
		return ctx.req
	}
	if val, ok := gcontext.GetOk(ctx.req, key); ok {
		return val
	}
	return ctx.Context.Value(key)
}

// ValueWithDefault returns the context value for the given key. If there's no
// such value, the provided default is returned.
func (ctx *requestContext) ValueWithDefault(key, defaultVal interface{}) interface{} {
	value := ctx.Value(key)
	if value == nil {
		value = defaultVal
	}
	return value
}

// ResponseFormat returns the response format for the request, defaulting to "json"
// if one is not specified using the "format" query parameter.
func (ctx *requestContext) ResponseFormat() string {
	return ctx.ValueWithDefault(FormatKey, "json").(string)
}

// ResourceId returns the resource id for the request, defaulting to an empty string
// if there isn't one.
func (ctx *requestContext) ResourceId() string {
	return ctx.ValueWithDefault(ResourceIdKey, "").(string)
}

// Version returns the API version for the request, defaulting to an empty string
// if one is not specified in the request path.
func (ctx *requestContext) Version() string {
	return ctx.ValueWithDefault(VersionKey, "").(string)
}

// HTTPRequest returns the *http.Request associated with ctx using NewContext,
// if any.
func HTTPRequest(ctx context.Context) (*http.Request, bool) {
	// We cannot use ctx.(*requestContext).req to get the request because ctx may
	// be a Context derived from a *requestContext. Instead, we use Value to
	// access the request if it is anywhere up the Context tree.
	req, ok := ctx.Value(RequestKey).(*http.Request)
	return req, ok
}
