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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
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

	// limitKey is the name of the query string variable for the results limit.
	limitKey = "limit"

	requestKey int = iota
	statusKey
	errorKey
	resultKey
)

// RequestContext contains the context information for the current HTTP request. It's a wrapper
// around Google's Context (http://godoc.org/code.google.com/p/go.net/context), which provides
// facilities for sending request-scoped values, cancelation signals, and deadlines
// across API boundaries to all the goroutines involved in handling a request.
type RequestContext interface {
	context.Context

	// WithValue returns a new RequestContext with the provided key-value pair and this context
	// as the parent.
	WithValue(interface{}, interface{}) RequestContext

	// ValueWithDefault returns the context value for the given key. If there's no such value,
	// the provided default is returned.
	ValueWithDefault(interface{}, interface{}) interface{}

	// Request returns the *http.Request associated with context using NewContext, if any.
	Request() (*http.Request, bool)

	// NextURL returns the URL to use to request the next page of results using the current
	// cursor. If there is no cursor for this request or the URL fails to be built, an empty
	// string is returned with the error set.
	NextURL() (string, error)

	// BuildURL builds a url.URL struct for a resource name & method.
	//
	// resourceName should have the same value as the handler's ResourceName method.
	//
	// method is the HandleMethod constant that corresponds with the resource
	// method for which to build the URL. E.g. HandleCreate with build a URL that
	// corresponds with the CreateResource method.
	//
	// All URL variables should be named in the vars map.
	BuildURL(resourceName string, method HandleMethod, vars RouteVars) (*url.URL, error)

	// ResponseFormat returns the response format for the request, defaulting to "json" if
	// one is not specified using the "format" query parameter.
	ResponseFormat() string

	// ResourceID returns the resource id for the request, defaulting to an empty string if
	// there isn't one.
	ResourceID() string

	// Version returns the API version for the request, defaulting to an empty string if
	// one is not specified in the request path.
	Version() string

	// Status returns the current HTTP status code that will be returned for the request,
	// defaulting to 200 if one hasn't been set yet.
	Status() int

	// setStatus sets the HTTP status code to be returned for the request.
	setStatus(int) RequestContext

	// Error returns the current error for the request or nil if no errors have been set.
	Error() error

	// setError sets the current error for the request.
	setError(error) RequestContext

	// Result returns the result resource for the request or nil if no result has been set.
	Result() interface{}

	// setResult sets the result resource for the request.
	setResult(interface{}) RequestContext

	// Cursor returns the current result cursor for the request, defaulting to an empty
	// string if one hasn't been set.
	Cursor() string

	// setCursor sets the current result cursor for the request.
	setCursor(string) RequestContext

	// Limit returns the maximum number of results that should be fetched.
	Limit() int

	// Messages returns all of the messages set by the request handler to be included in
	// the response.
	Messages() []string

	// AddMessage adds a message to the request messages to be included in the response.
	AddMessage(string)

	// Header returns the header key-value pairs for the request.
	Header() http.Header

	// Body returns a buffer containing the raw body of the request.
	Body() *bytes.Buffer

	// ResponseWriter Access to Response Writer Interface to allow for setting Response Header values
	ResponseWriter() http.ResponseWriter
}

// gorillaRequestContext is an implementation of the RequestContext interface. It wraps
// Gorilla's context package from which it attempts to store/retrieve values, delegating
// to the parent Context if necessary.
type gorillaRequestContext struct {
	context.Context
	req      *http.Request
	body     *bytes.Buffer
	writer   http.ResponseWriter
	router   *mux.Router
	messages []string
}

// NewContext returns a RequestContext populated with parameters from the request path and
// query string.
func NewContext(parent context.Context, req *http.Request, writer http.ResponseWriter) RequestContext {
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

	var body []byte
	if req.Body != nil {
		bytes, err := ioutil.ReadAll(req.Body)
		if err == nil {
			body = bytes
		}
	}

	// TODO: Keys can potentially be overwritten if the request path has
	// parameters with the same name as query string values. Figure out a
	// better way to handle this.

	return &gorillaRequestContext{parent, req, bytes.NewBuffer(body), writer, nil, []string{}}
}

func NewContextWithRouter(parent context.Context, req *http.Request, writer http.ResponseWriter,
	router *mux.Router) RequestContext {

	context := NewContext(parent, req, writer)
	context.(*gorillaRequestContext).router = router
	return context
}

// WithValue returns a new RequestContext with the provided key-value pair and this context
// as the parent.
func (ctx *gorillaRequestContext) WithValue(key, value interface{}) RequestContext {
	if r, ok := ctx.Request(); ok {
		return &gorillaRequestContext{context.WithValue(ctx, key, value), r, ctx.body, ctx.writer, ctx.router, ctx.messages}
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

// setStatus sets the HTTP status code to be returned for the request.
func (ctx *gorillaRequestContext) setStatus(status int) RequestContext {
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

// setError sets the current error for the request.
func (ctx *gorillaRequestContext) setError(err error) RequestContext {
	return ctx.WithValue(errorKey, err)
}

// Result returns the result resource for the request or nil if no result has been set.
func (ctx *gorillaRequestContext) Result() interface{} {
	return ctx.ValueWithDefault(resultKey, nil)
}

// setResult sets the result resource for the request.
func (ctx *gorillaRequestContext) setResult(result interface{}) RequestContext {
	return ctx.WithValue(resultKey, result)
}

// Cursor returns the current result cursor for the request, defaulting to an empty
// string if one hasn't been set.
func (ctx *gorillaRequestContext) Cursor() string {
	return ctx.ValueWithDefault(cursorKey, "").(string)
}

// setCursor sets the current result cursor for the request.
func (ctx *gorillaRequestContext) setCursor(cursor string) RequestContext {
	return ctx.WithValue(cursorKey, cursor)
}

// Header returns the header key-value pairs for the request.
func (ctx *gorillaRequestContext) Header() http.Header {
	req, ok := ctx.Request()
	if !ok {
		return http.Header{}
	}

	return req.Header
}

// Body returns a buffer containing the raw body of the request.
func (ctx *gorillaRequestContext) Body() *bytes.Buffer {
	return ctx.body
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
	limitStr := ctx.ValueWithDefault(limitKey, "100")
	limit, err := strconv.Atoi(limitStr.(string))
	if err != nil {
		limit = 100
	}
	return limit
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

// RouteVars is a map of URL route variables to values.
//
//     vars = RouteVars{"category": "widgets", "resource_id": "42"}
//
// Variables are defined in CreateURI and the other URI methods.
type RouteVars map[string]string

// BuildURL builds a full URL for a resource name & method.
//
// resourceName should have the same value as the handler's ResourceName method.
//
// method is the HandleMethod constant that corresponds with the resource
// method for which to build the URL. E.g. HandleCreate with build a URL that
// corresponds with the CreateResource method.
//
// All URL variables should be named in the vars map.
func (ctx *gorillaRequestContext) BuildURL(resourceName string,
	method HandleMethod, vars RouteVars) (*url.URL, error) {
	r, ok := ctx.Request()
	if !ok {
		return nil, fmt.Errorf("unable to build URL for resource name %q: no request available",
			resourceName)
	}

	routeName := resourceName + ":" + string(method)
	route := ctx.router.Get(routeName)

	// Transform RouteVars map to list of key, val pairs for Gorilla's API
	pairs := make([]string, (len(vars)*2)+2)
	for key, val := range vars {
		pairs = append(pairs, key, val)
	}
	pairs = append(pairs, "version", ctx.Version())
	url, err := route.URL(pairs...)
	if err != nil {
		return nil, err
	}
	url.Host = r.Host

	url.Scheme = "http"
	if r.TLS != nil {
		url.Scheme += "s"
	}

	return url, nil
}

// Messages returns all of the messages set by the request handler to be included in
// the response.
func (ctx *gorillaRequestContext) Messages() []string {
	messages := ctx.messages
	if err := ctx.Error(); err != nil {
		messages = append(messages, err.Error())
	}
	return messages
}

// AddMessage adds a message to the request messages to be included in the response.
func (ctx *gorillaRequestContext) AddMessage(message string) {
	ctx.messages = append(ctx.messages, message)
}

func (ctx *gorillaRequestContext) ResponseWriter() http.ResponseWriter {
	return ctx.writer
}
