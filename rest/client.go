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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Supported HTTP Methods
const (
	httpDelete = "DELETE"
	httpGet    = "GET"
	httpPost   = "POST"
	httpPut    = "PUT"
)

// InvocationHandler is a function that is to be wrapped by the ClientMiddleware
type InvocationHandler func(c *http.Client, method, url string, body interface{}, header http.Header) (*Response, error)

// ClientMiddleware is a function that wraps another function, and returns the wrapped function
type ClientMiddleware func(InvocationHandler) InvocationHandler

// HttpClient is the type that is used to perform HTTP Methods
type HttpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
	Get(url string) (resp *http.Response, err error)
	Post(url string, bodyType string, body io.Reader) (resp *http.Response, err error)
	PostForm(url string, data url.Values) (resp *http.Response, err error)
	Head(url string) (resp *http.Response, err error)
}

// RestClient performs HTTP methods including Get, Post, Put, and Delete
type RestClient interface {
	// Get will perform an HTTP GET on the specified URL and return the response.
	Get(url string, header http.Header) (*Response, error)

	// Post will perform an HTTP POST on the specified URL and return the response.
	Post(url string, body interface{}, header http.Header) (*Response, error)

	// Put will perform an HTTP PUT on the specified URL and return the response.
	Put(url string, body interface{}, header http.Header) (*Response, error)

	// Delete will perform an HTTP DELETE on the specified URL and return the response.
	Delete(url string, header http.Header) (*Response, error)
}

func NewRestClient(c HttpClient, middleware ...ClientMiddleware) RestClient {
	return &client{c, middleware}
}

// Client is the type that encapsulates and uses the Authorizer to sign any REST
// requests that are performed.
type Client struct {
	HttpClient
}

// client is the type that encapsulates and uses the Authorizer to sign any REST
// requests that are performed and has a list of middlewares to be applied to
// it's GET, POST, PUT, and DELETE functions.
type client struct {
	HttpClient
	middleware []ClientMiddleware
}

// Response is unmarshaled struct returned from an HTTP request.
type Response struct {
	Status   int            // HTTP status code.
	Reason   string         // Reason message for the status code.
	Messages []string       // Any server messages attached to the Response.
	Next     string         // A cursor to the next result set.
	Result   interface{}    // The decoded result of the REST request.
	Raw      *http.Response // The raw HTTP response.
}

// Wraps response decoding error in a helpful way
type ResponseDecodeError struct {
	StatusCode  int    // Response status code
	Status      string // Response status message
	Response    []byte // Payload of the response that could not be decoded
	DecodeError error  // Error that occurred while decoding the response
}

func (rde *ResponseDecodeError) Error() string {
	return fmt.Sprintf("(Error, Status Code, Status Message, Payload) = (%s, %d, %s, %s)",
		rde.DecodeError.Error(), rde.StatusCode, rde.Status, string(rde.Response))
}

// applyMiddleware wraps a given InvocationHandler with all of the middleware in the client
func (c *client) applyMiddleware(method InvocationHandler) InvocationHandler {
	for _, middleware := range c.middleware {
		method = middleware(method)
	}
	return method
}

func (c *client) process(method, url string, body interface{}, header http.Header) (*Response, error) {
	m := c.applyMiddleware(do)
	return m(c.HttpClient.(*http.Client), method, url, body, header)
}

// Get will perform an HTTP GET on the specified URL and return the response.
func (c *client) Get(url string, header http.Header) (*Response, error) {
	return c.process(httpGet, url, nil, header)
}

// Post will perform an HTTP POST on the specified URL and return the response.
func (c *client) Post(url string, body interface{}, header http.Header) (*Response, error) {
	return c.process(httpPost, url, body, header)
}

// Put will perform an HTTP PUT on the specified URL and return the response.
func (c *client) Put(url string, body interface{}, header http.Header) (*Response, error) {
	return c.process(httpPut, url, body, header)
}

// Delete will perform an HTTP DELETE on the specified URL and return the response.
func (c *client) Delete(url string, header http.Header) (*Response, error) {
	return c.process(httpDelete, url, nil, header)
}

/* Internal client calls. Would like to move up to the top level with a major
version change. Keeping this around for now to maintain backwards compat and allow
consumers to still construct the Client struct directly vs using the new constructor
method that returns the internal implementation that maps to the interface.
*/

// Get will perform an HTTP GET on the specified URL and return the response.
func (c *Client) Get(url string, header http.Header) (*Response, error) {
	return do(c.HttpClient.(*http.Client), httpGet, url, nil, header)
}

// Post will perform an HTTP POST on the specified URL and return the response.
func (c *Client) Post(url string, body interface{}, header http.Header) (*Response, error) {
	return do(c.HttpClient.(*http.Client), httpPost, url, body, header)
}

// Put will perform an HTTP PUT on the specified URL and return the response.
func (c *Client) Put(url string, body interface{}, header http.Header) (*Response, error) {
	return do(c.HttpClient.(*http.Client), httpPut, url, body, header)
}

// Delete will perform an HTTP DELETE on the specified URL and return the response.
func (c *Client) Delete(url string, header http.Header) (*Response, error) {
	return do(c.HttpClient.(*http.Client), httpDelete, url, nil, header)
}

var do = func(c *http.Client, method, url string, body interface{}, header http.Header) (*Response, error) {
	if header == nil {
		header = http.Header{}
	}

	var reqBody io.Reader
	switch method {
	case httpPost, httpPut:
		body, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(body)
		header.Set("Content-Type", "application/json")
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header = header

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Don't try to decode the response on 404.
	if resp.StatusCode == http.StatusNotFound {
		return &Response{
			Status:   resp.StatusCode,
			Reason:   http.StatusText(resp.StatusCode),
			Messages: []string{},
			Next:     "",
			Raw:      resp,
		}, nil
	}

	rawResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return decodeResponse(rawResp, resp)
}

func decodeResponse(response []byte, r *http.Response) (*Response, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(response, &payload); err != nil {
		return nil, &ResponseDecodeError{
			StatusCode:  r.StatusCode,
			Status:      r.Status,
			DecodeError: err,
			Response:    response,
		}
	}

	messages := []string{}
	for _, message := range payload["messages"].([]interface{}) {
		messages = append(messages, message.(string))
	}

	next := ""
	if n, ok := payload["next"]; ok {
		next = n.(string)
	}

	result, ok := payload["result"]
	if !ok {
		result = payload["results"]
	}

	resp := &Response{
		Status:   int(payload["status"].(float64)),
		Reason:   payload["reason"].(string),
		Messages: messages,
		Next:     next,
		Result:   result,
		Raw:      r,
	}

	return resp, nil
}
