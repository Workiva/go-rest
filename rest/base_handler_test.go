/*
Copyright 2014 Workiva, LLC

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
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestDefaultHandler struct {
	BaseResourceHandler
}

func (t TestDefaultHandler) ResourceName() string {
	return "foo"
}

// Ensures that CreateURI falls back to the correct default.
func TestCreateURIDefault(t *testing.T) {
	assert := assert.New(t)
	proxy := resourceHandlerProxy{TestDefaultHandler{}}

	assert.Equal("/api/v{version:[^/]+}/foo", proxy.CreateURI())
}

// Ensures that ReadURI falls back to the correct default.
func TestReadURIDefault(t *testing.T) {
	assert := assert.New(t)
	proxy := resourceHandlerProxy{TestDefaultHandler{}}

	assert.Equal("/api/v{version:[^/]+}/foo/{resource_id}", proxy.ReadURI())
}

// Ensures that ReadListURI falls back to the correct default.
func TestReadListURIDefault(t *testing.T) {
	assert := assert.New(t)
	proxy := resourceHandlerProxy{TestDefaultHandler{}}

	assert.Equal("/api/v{version:[^/]+}/foo", proxy.ReadListURI())
}

// Ensures that UpdateURI falls back to the correct default.
func TestUpdateURIDefault(t *testing.T) {
	assert := assert.New(t)
	proxy := resourceHandlerProxy{TestDefaultHandler{}}

	assert.Equal("/api/v{version:[^/]+}/foo/{resource_id}", proxy.UpdateURI())
}

// Ensures that DeleteURI falls back to the correct default.
func TestDeleteURIDefault(t *testing.T) {
	assert := assert.New(t)
	proxy := resourceHandlerProxy{TestDefaultHandler{}}

	assert.Equal("/api/v{version:[^/]+}/foo/{resource_id}", proxy.DeleteURI())
}

type TestHandler struct {
	BaseResourceHandler
}

func (t TestHandler) ResourceName() string {
	return "foo"
}

func (t TestHandler) CreateURI() string {
	return "/api/{version}/create_foo"
}

func (t TestHandler) ReadURI() string {
	return "/api/{version}/read_foo/{resource_id}"
}

func (t TestHandler) ReadListURI() string {
	return "/api/{version}/read_foo"
}

func (t TestHandler) UpdateURI() string {
	return "/api/{version}/update_foo/{resource_id}"
}

func (t TestHandler) DeleteURI() string {
	return "/api/{version}/delete_foo/{resource_id}"
}

// Ensures that CreateURI returns the custom URI.
func TestCreateURICustom(t *testing.T) {
	assert := assert.New(t)
	proxy := resourceHandlerProxy{TestHandler{}}

	assert.Equal("/api/{version}/create_foo", proxy.CreateURI())
}

// Ensures that ReadURI returns the custom URI.
func TestReadURICustom(t *testing.T) {
	assert := assert.New(t)
	proxy := resourceHandlerProxy{TestHandler{}}

	assert.Equal("/api/{version}/read_foo/{resource_id}", proxy.ReadURI())
}

// Ensures that ReadListURI returns the custom URI.
func TestReadListURICustom(t *testing.T) {
	assert := assert.New(t)
	proxy := resourceHandlerProxy{TestHandler{}}

	assert.Equal("/api/{version}/read_foo", proxy.ReadListURI())
}

// Ensures that UpdateURI returns the custom URI.
func TestUpdateURICustom(t *testing.T) {
	assert := assert.New(t)
	proxy := resourceHandlerProxy{TestHandler{}}

	assert.Equal("/api/{version}/update_foo/{resource_id}", proxy.UpdateURI())
}

// Ensures that DeleteURI returns the custom URI.
func TestDeleteURICustom(t *testing.T) {
	assert := assert.New(t)
	proxy := resourceHandlerProxy{TestHandler{}}

	assert.Equal("/api/{version}/delete_foo/{resource_id}", proxy.DeleteURI())
}
