package rest

import (
	"fmt"
	"net/http"
	"net/url"

	"code.google.com/p/go.net/context"
	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
)

const (
	// resourceIDKey is the name of the URL path variable for a resource ID.
	resourceIDKey = "resource_id"

	// formatKey is the name of the query string variable for the response format.
	formatKey = "format"

	// versionKey is the name of the URL path variable for the endpoint version.
	versionKey = "version"

	// cursorKey is the name of the query string variable for the results cursor.
	cursorKey = "next"

	requestKey int = iota
	statusKey
	errorKey
	resultKey
	limitKey
)

// RequestContext contains the context information for the current HTTP request. It's a wrapper
// around Google's Context (http://godoc.org/code.google.com/p/go.net/context), which provides
// facilities for sending request-scoped values, cancelation signals, and deadlines
// across API boundaries to all the goroutines involved in handling a request.
type RequestContext interface {
	context.Context
	WithValue(key, value interface{}) RequestContext
	ValueWithDefault(key, defaultVal interface{}) interface{}
	Request() (*http.Request, bool)
	NextURL() (string, error)
	ResponseFormat() string
	ResourceID() string
	Version() string
	Status() int
	SetStatus(int) RequestContext
	Error() error
	SetError(error) RequestContext
	Result() interface{}
	SetResult(interface{}) RequestContext
	Cursor() string
	SetCursor(string) RequestContext
	Limit() int
	SetLimit(int) RequestContext
}

// gorillaRequestContext is an implementation of the RequestContext interface. It wraps
// Gorilla's context package from which it attempts to store/retrieve values, delegating
// to the parent Context if necessary.
type gorillaRequestContext struct {
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

	return &gorillaRequestContext{parent, req}
}

// WithValue returns a new RequestContext with the provided key-value pair and this context
// as the parent.
func (ctx *gorillaRequestContext) WithValue(key, value interface{}) RequestContext {
	if r, ok := ctx.Request(); ok {
		return &gorillaRequestContext{context.WithValue(ctx, key, value), r}
	}

	// Should not reach this.
	panic("Unable to set value on context: no request")
}

// Value returns Gorilla's context package's value for this Context's request
// and key. It delegates to the parent Context if there is no such value.
func (ctx *gorillaRequestContext) Value(key interface{}) interface{} {
	if key == requestKey {
		return ctx.req
	}
	if val, ok := gcontext.GetOk(ctx.req, key); ok {
		return val
	}
	return ctx.Context.Value(key)
}

// ValueWithDefault returns the context value for the given key. If there's no
// such value, the provided default is returned.
func (ctx *gorillaRequestContext) ValueWithDefault(key, defaultVal interface{}) interface{} {
	value := ctx.Value(key)
	if value == nil {
		value = defaultVal
	}
	return value
}

// ResponseFormat returns the response format for the request, defaulting to "json"
// if one is not specified using the "format" query parameter.
func (ctx *gorillaRequestContext) ResponseFormat() string {
	return ctx.ValueWithDefault(formatKey, "json").(string)
}

// ResourceID returns the resource id for the request, defaulting to an empty string
// if there isn't one.
func (ctx *gorillaRequestContext) ResourceID() string {
	return ctx.ValueWithDefault(resourceIDKey, "").(string)
}

// Version returns the API version for the request, defaulting to an empty string
// if one is not specified in the request path.
func (ctx *gorillaRequestContext) Version() string {
	return ctx.ValueWithDefault(versionKey, "").(string)
}

// Status returns the current HTTP status code that will be returned for the request,
// defaulting to 200 if one hasn't been set yet.
func (ctx *gorillaRequestContext) Status() int {
	return ctx.ValueWithDefault(statusKey, http.StatusOK).(int)
}

// SetStatus sets the HTTP status code to be returned for the request.
func (ctx *gorillaRequestContext) SetStatus(status int) RequestContext {
	return ctx.WithValue(statusKey, status)
}

// Error returns the current error for the request or nil if no errors have been set.
func (ctx *gorillaRequestContext) Error() error {
	err := ctx.ValueWithDefault(errorKey, nil)

	if err == nil {
		return nil
	}

	return err.(error)
}

// SetError sets the current error for the request.
func (ctx *gorillaRequestContext) SetError(err error) RequestContext {
	return ctx.WithValue(errorKey, err)
}

// Result returns the result resource for the request or nil if no result has been set.
func (ctx *gorillaRequestContext) Result() interface{} {
	return ctx.ValueWithDefault(resultKey, nil)
}

// SetResult sets the result resource for the request.
func (ctx *gorillaRequestContext) SetResult(result interface{}) RequestContext {
	return ctx.WithValue(resultKey, result)
}

// Cursor returns the current result cursor for the request, defaulting to an empty
// string if one hasn't been set.
func (ctx *gorillaRequestContext) Cursor() string {
	return ctx.ValueWithDefault(cursorKey, "").(string)
}

// SetCursor sets the current result cursor for the request.
func (ctx *gorillaRequestContext) SetCursor(cursor string) RequestContext {
	return ctx.WithValue(cursorKey, cursor)
}

// Request returns the *http.Request associated with context using NewContext, if any.
func (ctx *gorillaRequestContext) Request() (*http.Request, bool) {
	// We cannot use ctx.(*gorillaRequestContext).req to get the request because ctx may
	// be a Context derived from a *gorillaRequestContext. Instead, we use Value to
	// access the request if it is anywhere up the Context tree.
	req, ok := ctx.Value(requestKey).(*http.Request)
	return req, ok
}

// Limit returns the maximum number of results that should be fetched.
func (ctx *gorillaRequestContext) Limit() int {
	return ctx.ValueWithDefault(limitKey, 100).(int)
}

// SetLimit sets the maximum number of results that should be fetched.
func (ctx *gorillaRequestContext) SetLimit(limit int) RequestContext {
	return ctx.WithValue(limitKey, limit)
}

// NextURL returns the URL to use to request the next page of results using the current
// cursor. If there is no cursor for this request or the URL fails to be built, an empty
// string is returned with the error set.
func (ctx *gorillaRequestContext) NextURL() (string, error) {
	cursor := ctx.Cursor()
	if cursor == "" {
		return "", fmt.Errorf("Unable to build next url: no cursor")
	}

	r, ok := ctx.Request()
	if !ok {
		return "", fmt.Errorf("Unable to build next url: no request")
	}

	var scheme string
	scheme = r.URL.Scheme
	if scheme == "" {
		scheme = "http"
	}

	urlStr := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("Unable to build next url: %s", urlStr)
	}

	q := u.Query()
	q.Set("next", cursor)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
