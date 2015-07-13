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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newMockDo(response *Response, err error) func(*http.Client, string, interface{}, http.Header) (*Response, error) {
	return func(*http.Client, string, interface{}, http.Header) (*Response, error) {
		return response, err
	}
}

// Ensures that Get invokes do with the correct HTTP client, method, url and
// header.
func TestClientGet(t *testing.T) {
	assert := assert.New(t)
	httpClient := http.DefaultClient
	client := &Client{httpClient}
	header := http.Header{}
	url := "http://localhost"
	mockResponse := &Response{}
	mockDo := func(c *http.Client, m, u string, b interface{}, h http.Header) (*Response, error) {
		assert.Equal(httpClient, c)
		assert.Equal(httpGet, m)
		assert.Nil(b)
		assert.Equal(header, h)
		assert.Equal(url, u)
		return mockResponse, nil
	}

	before := do
	do = mockDo

	resp, err := client.Get(url, header)

	assert.Equal(mockResponse, resp)
	assert.Nil(err)

	do = before
}

// Ensures that Post invokes do with the correct HTTP client, method, url and
// header.
func TestClientPost(t *testing.T) {
	assert := assert.New(t)
	httpClient := http.DefaultClient
	client := &Client{httpClient}
	header := http.Header{}
	url := "http://localhost"
	body := "foo"
	mockResponse := &Response{}
	mockDo := func(c *http.Client, m, u string, b interface{}, h http.Header) (*Response, error) {
		assert.Equal(httpClient, c)
		assert.Equal(httpPost, m)
		assert.Equal(body, b)
		assert.Equal(header, h)
		assert.Equal(url, u)
		return mockResponse, nil
	}

	before := do
	do = mockDo

	resp, err := client.Post(url, body, header)

	assert.Equal(mockResponse, resp)
	assert.Nil(err)

	do = before
}

// Ensures that Put invokes do with the correct HTTP client, method, url and
// header.
func TestClientPut(t *testing.T) {
	assert := assert.New(t)
	httpClient := http.DefaultClient
	client := &Client{httpClient}
	header := http.Header{}
	url := "http://localhost"
	body := "foo"
	mockResponse := &Response{}
	mockDo := func(c *http.Client, m, u string, b interface{}, h http.Header) (*Response, error) {
		assert.Equal(httpClient, c)
		assert.Equal(httpPut, m)
		assert.Equal(body, b)
		assert.Equal(header, h)
		assert.Equal(url, u)
		return mockResponse, nil
	}

	before := do
	do = mockDo

	resp, err := client.Put(url, body, header)

	assert.Equal(mockResponse, resp)
	assert.Nil(err)

	do = before
}

// Ensures that Delete invokes do with the correct HTTP client, method, url and
// header.
func TestClientDelete(t *testing.T) {
	assert := assert.New(t)
	httpClient := http.DefaultClient
	client := &Client{httpClient}
	header := http.Header{}
	url := "http://localhost"
	mockResponse := &Response{}
	mockDo := func(c *http.Client, m, u string, b interface{}, h http.Header) (*Response, error) {
		assert.Equal(httpClient, c)
		assert.Equal(httpDelete, m)
		assert.Nil(b)
		assert.Equal(header, h)
		assert.Equal(url, u)
		return mockResponse, nil
	}

	before := do
	do = mockDo

	resp, err := client.Delete(url, header)

	assert.Equal(mockResponse, resp)
	assert.Nil(err)

	do = before
}

// Ensures that do returns an error if the POST/PUT body is not JSON-
// marshalable.
func TestDoInvalidBody(t *testing.T) {
	var body interface{}
	_, err := do(http.DefaultClient, httpPost, "http://localhost", body, nil)
	assert.Error(t, err)
}

// Ensures that do returns an error if the request fails.
func TestDoBadRequest(t *testing.T) {
	_, err := do(http.DefaultClient, httpGet, "blah", nil, nil)
	assert.Error(t, err)
}

// Ensures that do returns a Response with a 404 code and the error is nil
// when a 404 response is received.
func TestDoNotFound(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	resp, err := do(http.DefaultClient, httpGet, ts.URL, nil, nil)

	if assert.NotNil(resp) {
		assert.Equal(http.StatusNotFound, resp.Status)
		assert.Equal(http.StatusText(http.StatusNotFound), resp.Reason)
		assert.Equal([]string{}, resp.Messages)
		assert.Equal("", resp.Next)
		assert.NotNil(resp.Raw)
		assert.Nil(resp.Result)
	}

	assert.Nil(err)
}

// Ensures that do returns an error when the response is not valid JSON.
func TestDoBadResponse(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"foo":`))
	}))
	defer ts.Close()

	_, err := do(http.DefaultClient, httpGet, ts.URL, nil, nil)

	assert.NotNil(err)
}

// Ensures that do returns a Response with the result decoded correctly for
// "list" endpoints ("results" key).
func TestDoDecodeResults(t *testing.T) {
	assert := assert.New(t)
	entity := map[string]interface{}{"foo": 1, "bar": "baz"}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := Payload{
			status:   http.StatusOK,
			reason:   http.StatusText(http.StatusOK),
			messages: []string{},
			results:  []interface{}{entity},
		}
		responseJSON, _ := json.Marshal(response)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	}))
	defer ts.Close()

	resp, err := do(http.DefaultClient, httpGet, ts.URL, nil, nil)

	if assert.NotNil(resp) {
		assert.Equal(http.StatusOK, resp.Status)
		assert.Equal(http.StatusText(http.StatusOK), resp.Reason)
		assert.Equal([]string{}, resp.Messages)
		assert.Equal("", resp.Next)
		assert.NotNil(resp.Raw)
		if assert.NotNil(resp.Result) {
			m := resp.Result.([]interface{})[0].(map[string]interface{})
			assert.Equal(float64(1.0), m["foo"])
			assert.Equal("baz", m["bar"])
		}
	}

	assert.Nil(err)
}

// Ensures that do returns a Response with the result decoded correctly for
// "non-list" endpoints ("result" key).
func TestDoDecodeResult(t *testing.T) {
	assert := assert.New(t)
	entity := map[string]interface{}{"foo": 1, "bar": "baz"}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := Payload{
			status:   http.StatusOK,
			reason:   http.StatusText(http.StatusOK),
			messages: []string{},
			result:   entity,
		}
		responseJSON, _ := json.Marshal(response)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	}))
	defer ts.Close()

	resp, err := do(http.DefaultClient, httpGet, ts.URL, nil, nil)

	if assert.NotNil(resp) {
		assert.Equal(http.StatusOK, resp.Status)
		assert.Equal(http.StatusText(http.StatusOK), resp.Reason)
		assert.Equal([]string{}, resp.Messages)
		assert.Equal("", resp.Next)
		assert.NotNil(resp.Raw)
		if assert.NotNil(resp.Result) {
			m := resp.Result.(map[string]interface{})
			assert.Equal(float64(1), m["foo"])
			assert.Equal("baz", m["bar"])
		}
	}

	assert.Nil(err)
}
