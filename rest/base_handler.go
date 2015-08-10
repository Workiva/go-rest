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
	"fmt"
	"net/http"
)

// BaseResourceHandler is a base implementation of ResourceHandler with stubs for the
// CRUD operations. This allows ResourceHandler implementations to only implement
// what they need.
type BaseResourceHandler struct{}

// ResourceName is a stub. It must be implemented.
func (b BaseResourceHandler) ResourceName() string {
	return ""
}

// CreateURI is a stub. Implement if necessary. The default create URI is
// /api/v{version:[^/]+}/resourceName.
func (b BaseResourceHandler) CreateURI() string {
	return ""
}

// CreateDocumentation is a stub. Implement if necessary.
func (b BaseResourceHandler) CreateDocumentation() string {
	return ""
}

// ReadURI is a stub. Implement if necessary. The default read URI is
// /api/v{version:[^/]+}/resourceName/{resource_id}.
func (b BaseResourceHandler) ReadURI() string {
	return ""
}

// ReadDocumentation is a stub. Implement if necessary.
func (b BaseResourceHandler) ReadDocumentation() string {
	return ""
}

// ReadListURI is a stub. Implement if necessary. The default read list URI is
// /api/v{version:[^/]+}/resourceName.
func (b BaseResourceHandler) ReadListURI() string {
	return ""
}

// ReadListDocumentation is a stub. Implement if necessary.
func (b BaseResourceHandler) ReadListDocumentation() string {
	return ""
}

// UpdateURI is a stub. Implement if necessary. The default update URI is
// /api/v{version:[^/]+}/resourceName/{resource_id}.
func (b BaseResourceHandler) UpdateURI() string {
	return ""
}

// UpdateDocumentation is a stub. Implement if necessary.
func (b BaseResourceHandler) UpdateDocumentation() string {
	return ""
}

// UpdateListURI is a stub. Implement if necessary. The default update list URI is
// /api/v{version:[^/]+}/resourceName.
func (b BaseResourceHandler) UpdateListURI() string {
	return ""
}

// UpdateListDocumentation is a stub. Implement if necessary.
func (b BaseResourceHandler) UpdateListDocumentation() string {
	return ""
}

// DeleteURI is a stub. Implement if necessary. The default delete URI is
// /api/v{version:[^/]+}/resourceName/{resource_id}.
func (b BaseResourceHandler) DeleteURI() string {
	return ""
}

// DeleteDocumentation is a stub. Implement if necessary.
func (b BaseResourceHandler) DeleteDocumentation() string {
	return ""
}

// CreateResource is a stub. Implement if necessary.
func (b BaseResourceHandler) CreateResource(ctx RequestContext, data Payload,
	version string) (Resource, error) {
	return nil, MethodNotAllowed("CreateResource is not implemented")
}

// ReadResourceList is a stub. Implement if necessary.
func (b BaseResourceHandler) ReadResourceList(ctx RequestContext, limit int,
	cursor string, version string) ([]Resource, string, error) {
	return nil, "", MethodNotAllowed("ReadResourceList not implemented")
}

// ReadResource is a stub. Implement if necessary.
func (b BaseResourceHandler) ReadResource(ctx RequestContext, id string,
	version string) (Resource, error) {
	return nil, MethodNotAllowed("ReadResource not implemented")
}

// UpdateResourceList is a stub. Implement if necessary.
func (b BaseResourceHandler) UpdateResourceList(ctx RequestContext, data []Payload,
	version string) ([]Resource, error) {
	return nil, MethodNotAllowed("UpdateResourceList not implemented")
}

// UpdateResource is a stub. Implement if necessary.
func (b BaseResourceHandler) UpdateResource(ctx RequestContext, id string,
	data Payload, version string) (Resource, error) {
	return nil, MethodNotAllowed("UpdateResource not implemented")
}

// DeleteResource is a stub. Implement if necessary.
func (b BaseResourceHandler) DeleteResource(ctx RequestContext, id string,
	version string) (Resource, error) {
	return nil, MethodNotAllowed("DeleteResource not implemented")
}

// Authenticate is the default authentication logic. All requests are authorized.
// Implement custom authentication logic if necessary.
func (b BaseResourceHandler) Authenticate(r *http.Request) error {
	return nil
}

func (b BaseResourceHandler) ValidVersions() []string {
	return nil
}

// Rules returns the resource rules to apply to incoming requests and outgoing
// responses. No rules are applied by default. Implement if necessary.
func (b BaseResourceHandler) Rules() Rules {
	return &rules{}
}

// resourceHandlerProxy wraps a ResourceHandler and allows the framework to provide
// additional logic around the proxied ResourceHandler, including default logic such
// as REST URIs.
type resourceHandlerProxy struct {
	ResourceHandler
}

// ResourceName returns the wrapped ResourceHandler's resource name. If the proxied
// handler doesn't have ResourceName implemented, it panics.
func (r resourceHandlerProxy) ResourceName() string {
	name := r.ResourceHandler.ResourceName()
	if name == "" {
		panic("ResourceHandler must implement ResourceName()")
	}
	return name
}

// CreateURI returns the URI for creating a resource using the handler-specified
// URI while falling back to a sensible default if not provided.
func (r resourceHandlerProxy) CreateURI() string {
	uri := r.ResourceHandler.CreateURI()
	if uri == "" {
		uri = fmt.Sprintf("/api/v{%s:[^/]+}/%s", versionKey, r.ResourceName())
	}
	return uri
}

// ReadURI returns the URI for reading a specific resource using the handler-specified
// URI while falling back to a sensible default if not provided.
func (r resourceHandlerProxy) ReadURI() string {
	uri := r.ResourceHandler.ReadURI()
	if uri == "" {
		uri = fmt.Sprintf("/api/v{%s:[^/]+}/%s/{%s}", versionKey, r.ResourceName(),
			resourceIDKey)
	}
	return uri
}

// ReadListURI returns the URI for reading a list of resources using the handler-
// specified URI while falling back to a sensible default if not provided.
func (r resourceHandlerProxy) ReadListURI() string {
	uri := r.ResourceHandler.ReadListURI()
	if uri == "" {
		uri = fmt.Sprintf("/api/v{%s:[^/]+}/%s", versionKey, r.ResourceName())
	}
	return uri
}

// UpdateURI returns the URI for updating a specific resource using the handler-
// specified URI while falling back to a sensible default if not provided.
func (r resourceHandlerProxy) UpdateURI() string {
	uri := r.ResourceHandler.UpdateURI()
	if uri == "" {
		uri = fmt.Sprintf("/api/v{%s:[^/]+}/%s/{%s}", versionKey, r.ResourceName(),
			resourceIDKey)
	}
	return uri
}

// UpdateListURI returns the URI for updating a list of resources using the handler-
// specified URI while falling back to a sensible default if not provided.
func (r resourceHandlerProxy) UpdateListURI() string {
	uri := r.ResourceHandler.UpdateListURI()
	if uri == "" {
		uri = fmt.Sprintf("/api/v{%s:[^/]+}/%s", versionKey, r.ResourceName())
	}
	return uri
}

// DeleteURI returns the URI for deleting a specific resource using the handler-
// specified URI while falling back to a sensible default if not provided.
func (r resourceHandlerProxy) DeleteURI() string {
	uri := r.ResourceHandler.DeleteURI()
	if uri == "" {
		uri = fmt.Sprintf("/api/v{%s:[^/]+}/%s/{%s}", versionKey,
			r.ResourceHandler.ResourceName(), resourceIDKey)
	}
	return uri
}
