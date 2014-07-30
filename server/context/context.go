package context

import (
	"net/http"

	"code.google.com/p/go.net/context"
	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
)

type RequestContext interface {
	context.Context
	ValueWithDefault(key, defaultVal interface{}) interface{}
	ResponseFormat() string
	ResourceId() string
	Version() string
}

// NewContext returns a Context whose Value method returns values associated
// with req using the Gorilla context package:
// http://www.gorillatoolkit.org/pkg/context
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

	return &wrapper{parent, req}
}

type wrapper struct {
	context.Context
	req *http.Request
}

const (
	RequestKey    = 0
	ResourceIdKey = "resource_id"
	FormatKey     = "format"
	VersionKey    = "version"
)

// Value returns Gorilla's context package's value for this Context's request
// and key. It delegates to the parent Context if there is no such value.
func (ctx *wrapper) Value(key interface{}) interface{} {
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
func (ctx *wrapper) ValueWithDefault(key, defaultVal interface{}) interface{} {
	value := ctx.Value(key)
	if value == nil {
		value = defaultVal
	}
	return value
}

func (ctx *wrapper) ResponseFormat() string {
	return ctx.ValueWithDefault(FormatKey, "json").(string)
}

func (ctx *wrapper) ResourceId() string {
	return ctx.ValueWithDefault(ResourceIdKey, "").(string)
}

func (ctx *wrapper) Version() string {
	return ctx.ValueWithDefault(VersionKey, "").(string)
}

// HTTPRequest returns the *http.Request associated with ctx using NewContext,
// if any.
func HTTPRequest(ctx context.Context) (*http.Request, bool) {
	// We cannot use ctx.(*wrapper).req to get the request because ctx may
	// be a Context derived from a *wrapper. Instead, we use Value to
	// access the request if it is anywhere up the Context tree.
	req, ok := ctx.Value(RequestKey).(*http.Request)
	return req, ok
}
