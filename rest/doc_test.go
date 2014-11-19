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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockContextGenerator struct {
	mock.Mock
}

func (m *mockContextGenerator) generate(handler ResourceHandler, version string) (
	map[string]interface{}, error) {

	args := m.Mock.Called(handler, version)
	ctx := args.Get(0)
	if ctx != nil {
		return ctx.(map[string]interface{}), args.Error(1)
	}
	return nil, args.Error(1)
}

type mockDocWriter struct {
	mock.Mock
}

func (m *mockDocWriter) mkdir(dir string, mode os.FileMode) error {
	args := m.Mock.Called(dir, mode)
	return args.Error(0)
}

func (m *mockDocWriter) write(file string, data []byte, mode os.FileMode) error {
	args := m.Mock.Called(file, data, mode)
	return args.Error(0)
}

type mockTemplateRenderer struct {
	mock.Mock
}

func (m *mockTemplateRenderer) render(context interface{}) string {
	args := m.Mock.Called(context)
	return args.String(0)
}

type mockTemplateParser struct {
	mock.Mock
}

func (m *mockTemplateParser) parse(template string) (templateRenderer, error) {
	args := m.Mock.Called(template)
	tpl := args.Get(0)
	if tpl != nil {
		return tpl.(templateRenderer), args.Error(1)
	}
	return nil, args.Error(1)
}

type fooResource struct {
	Foo string
	Bar int
	Baz []float32
	Qux time.Time
}

type fooHandler struct {
	BaseResourceHandler
}

func (f *fooHandler) ResourceName() string {
	return "foo"
}

func (f *fooHandler) CreateDocumentation() string {
	return "Creates a new foo"
}

func (f *fooHandler) ReadDocumentation() string {
	return "Retrieves a foo"
}

func (f *fooHandler) ReadListDocumentation() string {
	return "Retrieves a list of foos"
}

func (f *fooHandler) UpdateDocumentation() string {
	return "Updates a foo"
}

func (f *fooHandler) UpdateListDocumentation() string {
	return "Updates a list of foos"
}

func (f *fooHandler) Rules() Rules {
	return NewRules((*fooResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
			Type:       String,
			Versions:   []string{"1"},
			Required:   true,
			DocString:  "foo",
		},
		&Rule{
			Field:      "Bar",
			FieldAlias: "bar",
			Type:       Int,
			Versions:   []string{"1"},
			Required:   false,
			DocString:  "bar",
		},
		&Rule{
			Field:      "Baz",
			FieldAlias: "baz",
			Type:       Slice,
			Versions:   []string{"1"},
			Required:   true,
			DocString:  "baz",
		},
		&Rule{
			Field:      "Qux",
			FieldAlias: "qux",
			Type:       Time,
			Versions:   []string{"1"},
			Required:   true,
			DocString:  "qux",
		},
	)
}

type barResource struct {
	Foo time.Duration
	Bar bool
	Baz map[string]string
	Qux []fooResource
}

type barHandler struct {
	BaseResourceHandler
}

func (b *bazHandler) Rules() Rules {
	return NewRules((*fooResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
			Type:       String,
			Versions:   []string{"1"},
			Required:   true,
			DocString:  "foo",
		},
	)
}

func (b *barHandler) ResourceName() string {
	return "bar"
}

func (b *barHandler) CreateDocumentation() string {
	return "Creates a new bar"
}

func (b *barHandler) ReadDocumentation() string {
	return "Retrieves a bar"
}

func (b *barHandler) ReadListDocumentation() string {
	return "Retrieves a list of bars"
}

func (b *barHandler) UpdateDocumentation() string {
	return "Updates a bar"
}

func (b *barHandler) UpdateListDocumentation() string {
	return "Updates a list of bars"
}

func (b *barHandler) DeleteDocumentation() string {
	return "Deletes a bar"
}

func (b *barHandler) Rules() Rules {
	return NewRules((*barResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
			Type:       Duration,
			Versions:   []string{"1", "2"},
			Required:   true,
			DocString:  "foo",
		},
		&Rule{
			Field:      "Bar",
			FieldAlias: "bar",
			Type:       Bool,
			Versions:   []string{"1", "2"},
			Required:   false,
			DocString:  "bar",
		},
		&Rule{
			Field:      "Baz",
			FieldAlias: "baz",
			Type:       Map,
			Versions:   []string{"1", "2"},
			Required:   true,
			DocString:  "baz",
		},
		&Rule{
			Field:      "Qux",
			FieldAlias: "qux",
			Type:       Slice,
			Versions:   []string{"1", "2"},
			Required:   true,
			DocString:  "qux",
			Rules: NewRules((*fooResource)(nil),
				&Rule{
					Field:      "Foo",
					FieldAlias: "foo",
					Type:       String,
					Required:   true,
				},
				&Rule{
					Field:      "Bar",
					FieldAlias: "bar",
					Type:       Int,
					Required:   true,
				},
				&Rule{
					Field:      "Baz",
					FieldAlias: "baz",
					Type:       Slice,
					Required:   true,
				},
				&Rule{
					Field:      "Qux",
					FieldAlias: "qux",
					Type:       Time,
					Required:   true,
				},
			),
		},
	)
}

type bazHandler struct {
	BaseResourceHandler
}

func setupAPI() API {
	api := NewAPI(NewConfiguration())
	api.RegisterResourceHandler(&fooHandler{})
	api.RegisterResourceHandler(&barHandler{})
	return api
}

// Ensures that generateDocs returns an error when template parsing fails.
func TestGenerateDocsHandlesBadTemplate(t *testing.T) {
	assert := assert.New(t)
	api := setupAPI()
	mockParser := new(mockTemplateParser)
	mockContextGenerator := new(mockContextGenerator)
	mockDocWriter := new(mockDocWriter)
	mockParser.On("parse", indexTemplate).Return(nil, fmt.Errorf("error"))
	mockParser.On("parse", handlerTemplate).Return(nil, fmt.Errorf("error"))
	mockDocWriter.On("mkdir", "_docs/", 0777).Return(nil)
	docGenerator := &docGenerator{mockParser, mockContextGenerator, mockDocWriter}

	assert.NotNil(docGenerator.generateDocs(api), "Return value should not be nil")
}

// Ensures that generateDocs returns an error when the directory fails to be created.
func TestGenerateDocsHandlesDirectoryFail(t *testing.T) {
	assert := assert.New(t)
	api := setupAPI()
	mockParser := new(mockTemplateParser)
	mockContextGenerator := new(mockContextGenerator)
	mockDocWriter := new(mockDocWriter)
	mockDocWriter.On("mkdir", "_docs/", 0777).Return(fmt.Errorf("error"))
	docGenerator := &docGenerator{mockParser, mockContextGenerator, mockDocWriter}

	assert.NotNil(docGenerator.generateDocs(api), "Return value should not be nil")
}

// Ensures that generateDocs returns an error when writing an index file fails.
func TestGenerateDocsWriteFail(t *testing.T) {
	assert := assert.New(t)
	api := setupAPI()
	mockParser := new(mockTemplateParser)
	mockContextGenerator := new(mockContextGenerator)

	fooV1Context := map[string]interface{}{}
	barV1Context := map[string]interface{}{}
	barV2Context := map[string]interface{}{}
	mockContextGenerator.On("generate", api.ResourceHandlers()[0], "1").Return(fooV1Context, nil).Once()
	mockContextGenerator.On("generate", api.ResourceHandlers()[1], "1").Return(barV1Context, nil).Once()
	mockContextGenerator.On("generate", api.ResourceHandlers()[0], "2").Return(nil, fmt.Errorf("no endpoints")).Once()
	mockContextGenerator.On("generate", api.ResourceHandlers()[1], "2").Return(barV2Context, nil).Once()

	mockDocWriter := new(mockDocWriter)
	mockIndexTemplate := new(mockTemplateRenderer)
	mockHandlerTemplate := new(mockTemplateRenderer)

	mockParser.On("parse", indexTemplate).Return(mockIndexTemplate, nil)
	mockParser.On("parse", handlerTemplate).Return(mockHandlerTemplate, nil)

	fooV1Rendered := "foov1"
	barV1Rendered := "barv1"
	barV2Rendered := "barv2"
	mockHandlerTemplate.On("render", fooV1Context).Return(fooV1Rendered).Once()
	mockHandlerTemplate.On("render", barV1Context).Return(barV1Rendered).Once()
	mockHandlerTemplate.On("render", barV2Context).Return(barV2Rendered).Once()

	indexV1Rendered := "indexv1"
	indexV1Context := map[string]interface{}{
		"version":  "1",
		"versions": []string{"1", "2"},
		"handlers": []handlerDoc{
			map[string]string{"file": "fooresource_v1.html", "name": "fooResource"},
			map[string]string{"file": "barresource_v1.html", "name": "barResource"},
		},
	}
	indexV2Rendered := "indexv2"
	indexV2Context := map[string]interface{}{
		"version":  "2",
		"versions": []string{"1", "2"},
		"handlers": []handlerDoc{
			map[string]string{"file": "barresource_v2.html", "name": "barResource"},
		},
	}
	mockIndexTemplate.On("render", indexV1Context).Return(indexV1Rendered).Once()
	mockIndexTemplate.On("render", indexV2Context).Return(indexV2Rendered).Once()

	mockDocWriter.On("mkdir", "_docs/", 0777).Return(nil)
	mockDocWriter.On("write", "_docs/fooresource_v1.html", []byte(fooV1Rendered), 0644).Return(nil)
	mockDocWriter.On("write", "_docs/barresource_v1.html", []byte(barV1Rendered), 0644).Return(nil)
	mockDocWriter.On("write", "_docs/barresource_v2.html", []byte(barV2Rendered), 0644).Return(nil)
	mockDocWriter.On("write", "_docs/index_v1.html", []byte(indexV1Rendered), 0644).Return(fmt.Errorf("error"))

	docGenerator := &docGenerator{mockParser, mockContextGenerator, mockDocWriter}

	assert.NotNil(docGenerator.generateDocs(api), "Return value should be nil")
}

// Ensures that generateDocs returns an error when doc context generation fails.
func TestGenerateDocsGenerateFail(t *testing.T) {
	assert := assert.New(t)
	api := setupAPI()
	mockParser := new(mockTemplateParser)
	mockContextGenerator := new(mockContextGenerator)

	fooV1Context := map[string]interface{}{}
	mockContextGenerator.On("generate", api.ResourceHandlers()[0], "1").Return(fooV1Context, nil).Once()
	mockContextGenerator.On("generate", api.ResourceHandlers()[1], "1").Return(nil, fmt.Errorf("error")).Once()

	mockDocWriter := new(mockDocWriter)
	mockHandlerTemplate := new(mockTemplateRenderer)

	mockParser.On("parse", handlerTemplate).Return(mockHandlerTemplate, nil)

	fooV1Rendered := "foov1"
	mockHandlerTemplate.On("render", fooV1Context).Return(fooV1Rendered).Once()

	mockDocWriter.On("mkdir", "_docs/", 0777).Return(nil)
	mockDocWriter.On("write", "_docs/fooresource_v1.html", []byte(fooV1Rendered), 0644).Return(nil)

	docGenerator := &docGenerator{mockParser, mockContextGenerator, mockDocWriter}

	assert.NotNil(docGenerator.generateDocs(api), "Return value should not be nil")
}

// Ensures that generateDocs writes the correct data to the correct files.
func TestGenerateDocsHappyPath(t *testing.T) {
	assert := assert.New(t)
	api := setupAPI()
	mockParser := new(mockTemplateParser)
	mockContextGenerator := new(mockContextGenerator)

	fooV1Context := map[string]interface{}{}
	barV1Context := map[string]interface{}{}
	barV2Context := map[string]interface{}{}
	mockContextGenerator.On("generate", api.ResourceHandlers()[0], "1").Return(fooV1Context, nil).Once()
	mockContextGenerator.On("generate", api.ResourceHandlers()[1], "1").Return(barV1Context, nil).Once()
	mockContextGenerator.On("generate", api.ResourceHandlers()[0], "2").Return(nil, nil).Once()
	mockContextGenerator.On("generate", api.ResourceHandlers()[1], "2").Return(barV2Context, nil).Once()

	mockDocWriter := new(mockDocWriter)
	mockIndexTemplate := new(mockTemplateRenderer)
	mockHandlerTemplate := new(mockTemplateRenderer)

	mockParser.On("parse", indexTemplate).Return(mockIndexTemplate, nil)
	mockParser.On("parse", handlerTemplate).Return(mockHandlerTemplate, nil)

	fooV1Rendered := "foov1"
	barV1Rendered := "barv1"
	barV2Rendered := "barv2"
	mockHandlerTemplate.On("render", fooV1Context).Return(fooV1Rendered).Once()
	mockHandlerTemplate.On("render", barV1Context).Return(barV1Rendered).Once()
	mockHandlerTemplate.On("render", barV2Context).Return(barV2Rendered).Once()

	indexV1Rendered := "indexv1"
	indexV1Context := map[string]interface{}{
		"version":  "1",
		"versions": []string{"1", "2"},
		"handlers": []handlerDoc{
			map[string]string{"file": "fooresource_v1.html", "name": "fooResource"},
			map[string]string{"file": "barresource_v1.html", "name": "barResource"},
		},
	}
	indexV2Rendered := "indexv2"
	indexV2Context := map[string]interface{}{
		"version":  "2",
		"versions": []string{"1", "2"},
		"handlers": []handlerDoc{
			map[string]string{"file": "barresource_v2.html", "name": "barResource"},
		},
	}
	mockIndexTemplate.On("render", indexV1Context).Return(indexV1Rendered).Once()
	mockIndexTemplate.On("render", indexV2Context).Return(indexV2Rendered).Once()

	mockDocWriter.On("mkdir", "_docs/", 0777).Return(nil)
	mockDocWriter.On("write", "_docs/fooresource_v1.html", []byte(fooV1Rendered), 0644).Return(nil)
	mockDocWriter.On("write", "_docs/barresource_v1.html", []byte(barV1Rendered), 0644).Return(nil)
	mockDocWriter.On("write", "_docs/barresource_v2.html", []byte(barV2Rendered), 0644).Return(nil)
	mockDocWriter.On("write", "_docs/index_v1.html", []byte(indexV1Rendered), 0644).Return(nil)
	mockDocWriter.On("write", "_docs/index_v2.html", []byte(indexV2Rendered), 0644).Return(nil)

	docGenerator := &docGenerator{mockParser, mockContextGenerator, mockDocWriter}

	assert.Nil(docGenerator.generateDocs(api), "Return value should be nil")
}

// Ensures that generate returns nil context and nil error when there are no output fields for a
// version.
func TestGenerateNoOutput(t *testing.T) {
	assert := assert.New(t)
	generator := &defaultContextGenerator{}

	context, err := generator.generate(&resourceHandlerProxy{&fooHandler{}}, "2")

	assert.Nil(context, "Context should be nil")
	assert.Nil(err, "Error should be nil")
}

// Ensures that generate returns nil context and nil error when there are no documented endpoints.
func TestGenerateNoEndpoints(t *testing.T) {
	assert := assert.New(t)
	generator := &defaultContextGenerator{}

	context, err := generator.generate(&resourceHandlerProxy{&bazHandler{}}, "1")

	assert.Nil(context, "Context should be nil")
	assert.Nil(err, "Error should be nil")
}

// Ensures that generate returns the correct context for the handler version.
func TestGenerateHappyPath(t *testing.T) {
	assert := assert.New(t)
	generator := &defaultContextGenerator{}

	context, err := generator.generate(&resourceHandlerProxy{&fooHandler{}}, "1")

	if assert.NotNil(context, "Context should not be nil") {
		assert.Equal("fooResource", context["resource"])
		assert.Equal("1", context["version"])
		assert.Equal([]string{"1"}, context["versions"])
		assert.Equal("fooresource", context["fileNamePrefix"])
		endpoints := []endpoint{
			endpoint{
				"description":     "Creates a new foo",
				"exampleRequest":  "{\n    \"bar\": 0,\n    \"baz\": [],\n    \"foo\": \"foo\",\n    \"qux\": \"2014-09-05T15:45:36Z\"\n}",
				"exampleResponse": "{\n    \"bar\": 0,\n    \"baz\": [],\n    \"foo\": \"foo\",\n    \"qux\": \"2014-09-05T15:45:36Z\"\n}",
				"hasInput":        true,
				"index":           0,
				"label":           "success",
				"method":          "POST",
				"uri":             "/api/v1/foo",
				"inputFields": []field{
					field{
						"description": "foo",
						"name":        "foo",
						"required":    "required",
						"type":        "string",
					},
					field{
						"description": "bar",
						"name":        "bar",
						"required":    "optional",
						"type":        "int",
					},
					field{
						"description": "baz",
						"name":        "baz",
						"required":    "required",
						"type":        "[]interface{}",
					},
					field{
						"description": "qux",
						"name":        "qux",
						"required":    "required",
						"type":        "time.Time",
					},
				},
				"outputFields": []field{
					field{
						"description": "foo",
						"name":        "foo",
						"type":        "string",
					},
					field{
						"description": "bar",
						"name":        "bar",
						"type":        "int",
					},
					field{
						"description": "baz",
						"name":        "baz",
						"type":        "[]interface{}",
					},
					field{
						"description": "qux",
						"name":        "qux",
						"type":        "time.Time",
					},
				},
			},
			endpoint{
				"description":     "Retrieves a list of foos",
				"exampleResponse": "[\n    {\n        \"bar\": 0,\n        \"baz\": [],\n        \"foo\": \"foo\",\n        \"qux\": \"2014-09-05T15:45:36Z\"\n    }\n]",
				"hasInput":        false,
				"index":           1,
				"label":           "info",
				"method":          "GET",
				"uri":             "/api/v1/foo",
				"outputFields": []field{
					field{
						"description": "foo",
						"name":        "foo",
						"type":        "string",
					},
					field{
						"description": "bar",
						"name":        "bar",
						"type":        "int",
					},
					field{
						"description": "baz",
						"name":        "baz",
						"type":        "[]interface{}",
					},
					field{
						"description": "qux",
						"name":        "qux",
						"type":        "time.Time",
					},
				},
			},
			endpoint{
				"description":     "Retrieves a foo",
				"exampleResponse": "{\n    \"bar\": 0,\n    \"baz\": [],\n    \"foo\": \"foo\",\n    \"qux\": \"2014-09-05T15:45:36Z\"\n}",
				"hasInput":        false,
				"index":           2,
				"label":           "info",
				"method":          "GET",
				"uri":             "/api/v1/foo/:resource_id",
				"outputFields": []field{
					field{
						"description": "foo",
						"name":        "foo",
						"type":        "string",
					},
					field{
						"description": "bar",
						"name":        "bar",
						"type":        "int",
					},
					field{
						"description": "baz",
						"name":        "baz",
						"type":        "[]interface{}",
					},
					field{
						"description": "qux",
						"name":        "qux",
						"type":        "time.Time",
					},
				},
			},
			endpoint{
				"description":     "Updates a list of foos",
				"exampleRequest":  "[\n    {\n        \"bar\": 0,\n        \"baz\": [],\n        \"foo\": \"foo\",\n        \"qux\": \"2014-09-05T15:45:36Z\"\n    }\n]",
				"exampleResponse": "[\n    {\n        \"bar\": 0,\n        \"baz\": [],\n        \"foo\": \"foo\",\n        \"qux\": \"2014-09-05T15:45:36Z\"\n    }\n]",
				"hasInput":        true,
				"index":           3,
				"label":           "warning",
				"method":          "PUT",
				"uri":             "/api/v1/foo",
				"inputFields": []field{
					field{
						"description": "foo",
						"name":        "foo",
						"required":    "required",
						"type":        "string",
					},
					field{
						"description": "bar",
						"name":        "bar",
						"required":    "optional",
						"type":        "int",
					},
					field{
						"description": "baz",
						"name":        "baz",
						"required":    "required",
						"type":        "[]interface{}",
					},
					field{
						"description": "qux",
						"name":        "qux",
						"required":    "required",
						"type":        "time.Time",
					},
				},
				"outputFields": []field{
					field{
						"description": "foo",
						"name":        "foo",
						"type":        "string",
					},
					field{
						"description": "bar",
						"name":        "bar",
						"type":        "int",
					},
					field{
						"description": "baz",
						"name":        "baz",
						"type":        "[]interface{}",
					},
					field{
						"description": "qux",
						"name":        "qux",
						"type":        "time.Time",
					},
				},
			},
			endpoint{
				"description":     "Updates a foo",
				"exampleRequest":  "{\n    \"bar\": 0,\n    \"baz\": [],\n    \"foo\": \"foo\",\n    \"qux\": \"2014-09-05T15:45:36Z\"\n}",
				"exampleResponse": "{\n    \"bar\": 0,\n    \"baz\": [],\n    \"foo\": \"foo\",\n    \"qux\": \"2014-09-05T15:45:36Z\"\n}",
				"hasInput":        true,
				"index":           4,
				"label":           "warning",
				"method":          "PUT",
				"uri":             "/api/v1/foo/:resource_id",
				"inputFields": []field{
					field{
						"description": "foo",
						"name":        "foo",
						"required":    "required",
						"type":        "string",
					},
					field{
						"description": "bar",
						"name":        "bar",
						"required":    "optional",
						"type":        "int",
					},
					field{
						"description": "baz",
						"name":        "baz",
						"required":    "required",
						"type":        "[]interface{}",
					},
					field{
						"description": "qux",
						"name":        "qux",
						"required":    "required",
						"type":        "time.Time",
					},
				},
				"outputFields": []field{
					field{
						"description": "foo",
						"name":        "foo",
						"type":        "string",
					},
					field{
						"description": "bar",
						"name":        "bar",
						"type":        "int",
					},
					field{
						"description": "baz",
						"name":        "baz",
						"type":        "[]interface{}",
					},
					field{
						"description": "qux",
						"name":        "qux",
						"type":        "time.Time",
					},
				},
			},
		}
		for i, endpoint := range context["endpoints"].([]endpoint) {
			expected := endpoints[i]
			assert.Equal(expected["description"], endpoint["description"])
			assert.Equal(expected["exampleRequest"], endpoint["exampleRequest"])
			assert.Equal(expected["exampleResponse"], endpoint["exampleResponse"])
			assert.Equal(expected["hasInput"], endpoint["hasInput"])
			assert.Equal(expected["index"], endpoint["index"])
			assert.Equal(expected["inputFields"], endpoint["inputFields"])
			assert.Equal(expected["outputFields"], endpoint["outputFields"])
			assert.Equal(expected["label"], endpoint["label"])
			assert.Equal(expected["method"], endpoint["method"])
			assert.Equal(expected["uri"], endpoint["uri"])
		}
	}
	assert.Nil(err, "Error should be nil")
}
