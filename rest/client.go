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
)

// Supported HTTP Methods
const (
	httpDelete = "DELETE"
	httpGet    = "GET"
	httpPost   = "POST"
	httpPut    = "PUT"
)

// Client is the type that encapsulates and uses the Authorizer to sign any REST
// requests that are performed.
type Client struct {
	*http.Client
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

// Get will perform an HTTP GET on the specified URL and return the response.
func (c *Client) Get(url string, header http.Header) (*Response, error) {
	return do(c.Client, httpGet, url, nil, header)
}

// Post will perform an HTTP POST on the specified URL and return the response.
func (c *Client) Post(url string, body interface{}, header http.Header) (*Response, error) {
	return do(c.Client, httpPost, url, body, header)
}

// Put will perform an HTTP PUT on the specified URL and return the response.
func (c *Client) Put(url string, body interface{}, header http.Header) (*Response, error) {
	return do(c.Client, httpPut, url, body, header)
}

// Delete will perform an HTTP DELETE on the specified URL and return the response.
func (c *Client) Delete(url string, header http.Header) (*Response, error) {
	return do(c.Client, httpDelete, url, nil, header)
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
		return nil, fmt.Errorf("%v. Response with status %d: %s had unmarshallable payload %s", err, r.StatusCode, r.Status, string(response))
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
