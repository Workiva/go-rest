/*
Package rest contains the company standard REST API client and server implementations.

This package can be used with any type that implements the Consumer interface:

    rc := rest.Client{myConsumer{
    	"Key",
    	"Secret",
    }}

    params := map[string]string{
        "something": "cool"
    }

    resp, err := rc.GetJSON("http://example.com/api/", params, nil)
*/
package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Supported HTTP Methods
const (
	DELETE = "DELETE"
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
)

// Authorizer encapsulates the required authentication logic for the API the
// Client will interact with.
type Authorizer interface {
	Authorize(urlStr string, requestType string, form url.Values) url.Values
}

// Client is the type that encapsulates and uses the Consumer to sign any REST
// requests that are performed.
type Client struct {
	Authorizer
}

// BaseResponse is the resultant type of any of the Do*() methods of the
// Client. It contains several informational fields as well as the result
// value.
type BaseResponse struct {
	Status   int         // HTTP status code.
	Reason   string      // Reason message for the status code.
	Messages []string    // Any server messages attached to the Response.
	Next     string      // A cursor to the next result set.
	Results  interface{} // The actual results of the REST request.
}

func (c *Client) decode(r io.Reader, want interface{}) (*BaseResponse, error) {
	resp := &BaseResponse{Results: want}
	err := json.NewDecoder(r).Decode(resp)
	return resp, err
}

// BuildForm will take a map of the form input and build a url.Values object.
func (c *Client) BuildForm(params map[string]string) url.Values {
	form := url.Values{}

	for param, value := range params {
		form.Set(param, value)
	}

	return form
}

func (c *Client) do(method, urlStr string, params map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return nil, err
	}

	// Set up the form values and request method.
	req.Method = method
	form := c.BuildForm(params)

	// TODO: Find a way to push the encoding and authorization into the specific http methods themselves.
	switch method {
	case GET, DELETE:
		// Combine form and auth into the query string.
		req.URL.RawQuery = c.Authorize(urlStr, method, form).Encode()
	case POST, PUT:
		// Set the auth params in the query string.
		req.URL.RawQuery = c.Authorize(urlStr, method, url.Values{}).Encode()
		// Encode the form values as JSON and put them in the request body.
		body, err := encodeBody(form)
		if err != nil {
			return nil, err
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}

	return http.DefaultClient.Do(req)
}

func (c *Client) doJSON(method, url string, params map[string]string, entity interface{}) (*BaseResponse, error) {
	response, err := c.do(method, url, params)
	if err != nil {
		return &BaseResponse{}, err
	}

	defer response.Body.Close()

	b, _ := ioutil.ReadAll(response.Body)
	return c.decode(bytes.NewReader(b), entity)
}

// Get will perform a HTTP GET against the supplied URL with the given parameters.
func (c *Client) Get(urlStr string, params map[string]string) (*http.Response, error) {
	return c.do(GET, urlStr, params)
}

// GetJSON will perform a HTTP GET and will JSON decode the response.
func (c *Client) GetJSON(url string, params map[string]string, entity interface{}) (*BaseResponse, error) {
	return c.doJSON(GET, url, params, entity)
}

// Post will perform a HTTP POST against the supplied URL with the given parameters.
func (c *Client) Post(urlStr string, params map[string]string) (*http.Response, error) {
	return c.do(POST, urlStr, params)
}

// PostJSON will perform a HTTP POST and will JSON decode the response.
func (c *Client) PostJSON(url string, params map[string]string, entity interface{}) (*BaseResponse, error) {
	return c.doJSON(POST, url, params, entity)
}

// Put will perform a HTTP PUT against the supplied URL with the given parameters.
func (c *Client) Put(urlStr string, params map[string]string) (*http.Response, error) {
	return c.do(PUT, urlStr, params)
}

// PutJSON will perform a HTTP PUT and will JSON decode the response.
func (c *Client) PutJSON(url string, params map[string]string, entity interface{}) (*BaseResponse, error) {
	return c.doJSON(PUT, url, params, entity)
}

// Delete will perform a HTTP DELETE against the supplied URL with the given parameters.
func (c *Client) Delete(urlStr string, params map[string]string) (*http.Response, error) {
	return c.do(DELETE, urlStr, params)
}

// DeleteJSON will perform a HTTP DELETE and will JSON decode the response.
func (c *Client) DeleteJSON(url string, params map[string]string, entity interface{}) (*BaseResponse, error) {
	return c.doJSON(DELETE, url, params, entity)
}

func encodeBody(form url.Values) ([]byte, error) {
	actual := map[string]string{}

	// Get the first value for each key.
	for k := range form {
		actual[k] = form.Get(k)
	}

	// Encode the form
	body, err := json.Marshal(actual)
	if err != nil {
		return nil, err
	}
	return body, nil
}
