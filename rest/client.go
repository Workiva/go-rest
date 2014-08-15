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

    resp, err := rc.GetJson("http://example.com/api/", params, nil)
*/
package rest

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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

	form := c.Authorize(urlStr, method, c.BuildForm(params))

	// TODO: Investigate if this needs to be different for POST
	req.URL.RawQuery = form.Encode()

	return http.DefaultClient.Do(req)
}

func (c *Client) doJson(method, url string, params map[string]string, entity interface{}) (*BaseResponse, error) {
	response, err := c.do(method, url, params)
	if err != nil {
		return &BaseResponse{}, err
	}

	defer response.Body.Close()

	b, _ := ioutil.ReadAll(response.Body)
	return c.decode(strings.NewReader(string(b)), entity)
}

// Get will perform a HTTP GET against the supplied URL with the given parameters.
func (c *Client) Get(urlStr string, params map[string]string) (*http.Response, error) {
	return c.do(GET, urlStr, params)
}

// GetJson will perform a HTTP GET and will JSON decode the response.
func (c *Client) GetJson(url string, params map[string]string, entity interface{}) (*BaseResponse, error) {
	return c.doJson(GET, url, params, entity)
}

// Post will perform a HTTP POST against the supplied URL with the given parameters.
func (c *Client) Post(urlStr string, params map[string]string) (*http.Response, error) {
	return c.do(POST, urlStr, params)
}

// PostJson will perform a HTTP POST and will JSON decode the response.
func (c *Client) PostJson(url string, params map[string]string, entity interface{}) (*BaseResponse, error) {
	return c.doJson(POST, url, params, entity)
}

// Put will perform a HTTP PUT against the supplied URL with the given parameters.
func (c *Client) Put(urlStr string, params map[string]string) (*http.Response, error) {
	return c.do(PUT, urlStr, params)
}

// PutJson will perform a HTTP PUT and will JSON decode the response.
func (c *Client) PutJson(url string, params map[string]string, entity interface{}) (*BaseResponse, error) {
	return c.doJson(PUT, url, params, entity)
}

// Delete will perform a HTTP DELETE against the supplied URL with the given parameters.
func (c *Client) Delete(urlStr string, params map[string]string) (*http.Response, error) {
	return c.do(DELETE, urlStr, params)
}

// DeleteJson will perform a HTTP DELETE and will JSON decode the response.
func (c *Client) DeleteJson(url string, params map[string]string, entity interface{}) (*BaseResponse, error) {
	return c.doJson(DELETE, url, params, entity)
}
