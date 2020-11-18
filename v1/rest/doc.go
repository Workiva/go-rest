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
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/hoisie/mustache"
)

type endpoint map[string]interface{}
type field map[string]interface{}
type handlerDoc map[string]string

// templateRenderer is a template which can be rendered as a string.
type templateRenderer interface {
	// render will render the template as a string using the provided context values.
	render(interface{}) string
}

// templateParser is used to parse a template string into a templateRenderer struct.
type templateParser interface {
	// parse will parse the string into a templateRenderer or return an error if the
	// template is malformed.
	parse(string) (templateRenderer, error)
}

// mustacheRenderer is an implementation of the templateRenderer interface which relies on
// mustache templating.
type mustacheRenderer struct {
	*mustache.Template
}

// render will render the template as a string using the provided context values.
func (m *mustacheRenderer) render(context interface{}) string {
	return m.Render(context)
}

// mustacheParser is an implementation of the templateParser interface which relies on
// mustache templating.
type mustacheParser struct{}

// parse will parse the string into a templateRenderer or return an error if the template
// is malformed.
func (m *mustacheParser) parse(template string) (templateRenderer, error) {
	tpl, err := mustache.ParseString(template)
	if err != nil {
		return nil, err
	}

	return &mustacheRenderer{tpl}, nil
}

// docContextGenerator creates template contexts for rendering ResourceHandler documentation.
type docContextGenerator interface {
	// generate creates a template context for the provided ResourceHandler.
	generate(ResourceHandler, string) (map[string]interface{}, error)
}

// docWriter writes rendered documentation to a persistent medium.
type docWriter interface {
	// mkdir creates a directory to store documentation in.
	mkdir(string, os.FileMode) error

	// write saves the rendered documentation.
	write(string, []byte, os.FileMode) error
}

// fsDocWriter is an implementation of the docWriter interface which writes documentation to
// the local file system.
type fsDocWriter struct{}

// mkdir creates a directory to store documentation in.
func (f *fsDocWriter) mkdir(dir string, mode os.FileMode) error {
	return os.MkdirAll(dir, mode)
}

// write saves the rendered documentation.
func (f *fsDocWriter) write(file string, data []byte, mode os.FileMode) error {
	return ioutil.WriteFile(file, data, mode)
}

// docGenerator produces documentation files for APIs by introspecting ResourceHandlers and
// their Rules.
type docGenerator struct {
	templateParser
	docContextGenerator
	docWriter
}

// newDocGenerator creates a new docGenerator instance which relies on mustache templating.
func newDocGenerator() *docGenerator {
	return &docGenerator{
		&mustacheParser{},
		&defaultContextGenerator{},
		&fsDocWriter{},
	}
}

// generateDocs creates the HTML documentation for the provided API. The resulting HTML files
// will be placed in the directory specified by the API Configuration. Returns an error if
// generating the documentation failed, nil otherwise.
func (d *docGenerator) generateDocs(api API) error {
	dir := api.Configuration().DocsDirectory
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	if err := d.mkdir(dir, os.FileMode(0777)); err != nil {
		api.Configuration().Logger.Println(err)
		return err
	}

	handlers := api.ResourceHandlers()
	docs := map[string][]handlerDoc{}
	versions := versions(handlers)

	for _, version := range versions {
		versionDocs := make([]handlerDoc, 0, len(handlers))
		for _, handler := range handlers {
			doc, err := d.generateHandlerDoc(handler, version, dir)
			if err != nil {
				api.Configuration().Logger.Println(err)
				return err
			} else if doc != nil {
				versionDocs = append(versionDocs, doc)
			}
		}

		docs[version] = versionDocs
	}

	if err := d.generateIndexDocs(docs, versions, dir); err != nil {
		api.Configuration().Logger.Println(err)
		return err
	}

	api.Configuration().Debugf("Documentation generated in %s", dir)
	return nil
}

// generateIndexDocs creates index files for each API version with documented endpoints.
func (d *docGenerator) generateIndexDocs(docs map[string][]handlerDoc, versions []string,
	dir string) error {

	tpl, err := d.parse(indexTemplate)
	if err != nil {
		return err
	}

	for version, docList := range docs {
		rendered := tpl.render(map[string]interface{}{
			"handlers": docList,
			"version":  version,
			"versions": versions,
		})
		if err := d.write(fmt.Sprintf("%sindex_v%s.html", dir, version),
			[]byte(rendered), 0644); err != nil {
			return err
		}
	}

	return nil
}

// generateHandlerDoc creates a documentation file for the versioned ResourceHandler.
// Returns nil if the handler contains no documented endpoints or has no output fields.
func (d *docGenerator) generateHandlerDoc(handler ResourceHandler, version,
	dir string) (handlerDoc, error) {

	tpl, err := d.parse(handlerTemplate)
	if err != nil {
		return nil, err
	}

	context, err := d.generate(handler, version)
	if context == nil || err != nil {
		return nil, err
	}
	rendered := tpl.render(context)

	name := handlerTypeName(handler)
	file := fileName(name, version)
	if err := d.write(fmt.Sprintf("%s%s", dir, file), []byte(rendered), 0644); err != nil {
		return nil, err
	}

	doc := handlerDoc{"name": name, "file": file}
	return doc, nil
}

// defaultContextGenerator is an implementation of the docContextGenerator interface.
type defaultContextGenerator struct{}

// generate creates a template context for the provided ResourceHandler.
func (d *defaultContextGenerator) generate(handler ResourceHandler, version string) (
	map[string]interface{}, error) {

	inputFields := getInputFields(handler.Rules().ForVersion(version))
	outputFields := getOutputFields(handler.Rules().ForVersion(version))

	if len(inputFields) == 0 && len(outputFields) == 0 {
		// Handler has no fields for this version.
		return nil, nil
	}

	index := 0
	endpoints := []endpoint{}
	if handler.CreateDocumentation() != "" {
		endpoints = append(endpoints, endpoint{
			"uri":             formatURI(handler.CreateURI(), version),
			"method":          "POST",
			"label":           "success",
			"description":     handler.CreateDocumentation(),
			"hasInput":        true,
			"inputFields":     inputFields,
			"outputFields":    outputFields,
			"exampleRequest":  buildExampleRequest(handler.Rules(), false, version),
			"exampleResponse": buildExampleResponse(handler.Rules(), false, version),
			"index":           index,
		})
	}
	index++

	if handler.ReadListDocumentation() != "" {
		endpoints = append(endpoints, endpoint{
			"uri":             formatURI(handler.ReadListURI(), version),
			"method":          "GET",
			"label":           "info",
			"description":     handler.ReadListDocumentation(),
			"hasInput":        false,
			"outputFields":    outputFields,
			"exampleResponse": buildExampleResponse(handler.Rules(), true, version),
			"index":           index,
		})
	}
	index++

	if handler.ReadDocumentation() != "" {
		endpoints = append(endpoints, endpoint{
			"uri":             formatURI(handler.ReadURI(), version),
			"method":          "GET",
			"label":           "info",
			"description":     handler.ReadDocumentation(),
			"hasInput":        false,
			"outputFields":    outputFields,
			"exampleResponse": buildExampleResponse(handler.Rules(), false, version),
			"index":           index,
		})
	}
	index++

	if handler.UpdateListDocumentation() != "" {
		endpoints = append(endpoints, endpoint{
			"uri":             formatURI(handler.UpdateListURI(), version),
			"method":          "PUT",
			"label":           "warning",
			"description":     handler.UpdateListDocumentation(),
			"hasInput":        true,
			"inputFields":     inputFields,
			"outputFields":    outputFields,
			"exampleRequest":  buildExampleRequest(handler.Rules(), true, version),
			"exampleResponse": buildExampleResponse(handler.Rules(), true, version),
			"index":           index,
		})
	}
	index++

	if handler.UpdateDocumentation() != "" {
		endpoints = append(endpoints, endpoint{
			"uri":             formatURI(handler.UpdateURI(), version),
			"method":          "PUT",
			"label":           "warning",
			"description":     handler.UpdateDocumentation(),
			"hasInput":        true,
			"inputFields":     inputFields,
			"outputFields":    outputFields,
			"exampleRequest":  buildExampleRequest(handler.Rules(), false, version),
			"exampleResponse": buildExampleResponse(handler.Rules(), false, version),
			"index":           index,
		})
	}
	index++

	if handler.DeleteDocumentation() != "" {
		endpoints = append(endpoints, endpoint{
			"uri":             formatURI(handler.DeleteURI(), version),
			"method":          "DELETE",
			"label":           "danger",
			"description":     handler.DeleteDocumentation(),
			"hasInput":        false,
			"outputFields":    outputFields,
			"exampleResponse": buildExampleResponse(handler.Rules(), false, version),
			"index":           index,
		})
	}
	index++

	if len(endpoints) == 0 {
		// No documented endpoints.
		return nil, nil
	}

	name := handlerTypeName(handler)
	context := map[string]interface{}{
		"resource":       name,
		"version":        version,
		"versions":       handlerVersions(handler),
		"endpoints":      endpoints,
		"fileNamePrefix": fileNamePrefix(name),
	}

	return context, nil
}

// formatURI returns the specified URI replacing templated variable names with their
// human-readable documentation equivalent. It also replaces the version regex with
// the actual version string.
func formatURI(uri, version string) string {
	uri = strings.Replace(uri, "{version:[^/]+}", version, -1)

	re := regexp.MustCompile("{.*?}")

	for _, param := range re.FindAllString(uri, -1) {
		uri = replaceURIParam(uri, param)
	}

	return uri
}

// replaceURIParam replaces the templated variable name with the human-readable
// documentation equivalent, e.g. {foo} is replaced with :foo.
func replaceURIParam(uri, param string) string {
	paramName := param[1 : len(param)-1]
	return strings.Replace(uri, param, ":"+paramName, -1)
}

// getInputFields returns input field descriptions.
func getInputFields(rules Rules) []field {
	rules = rules.Filter(Inbound)
	fields := make([]field, 0, rules.Size())

	for _, rule := range rules.Contents() {
		required := "optional"
		if rule.Required {
			required = "required"
		}

		field := field{
			"name":        rule.Name(),
			"required":    required,
			"type":        ruleTypeName(rule, Inbound),
			"description": rule.DocString,
		}

		fields = append(fields, field)
	}

	return fields
}

// getInputFields returns output field descriptions.
func getOutputFields(rules Rules) []field {
	rules = rules.Filter(Outbound)
	fields := make([]field, 0, rules.Size())

	for _, rule := range rules.Contents() {
		field := field{
			"name":        rule.Name(),
			"type":        ruleTypeName(rule, Outbound),
			"description": rule.DocString,
		}

		fields = append(fields, field)
	}

	return fields
}

// ruleTypeName returns the human-readable type name for a Rule.
func ruleTypeName(r *Rule, filter Filter) string {
	name := typeToName[r.Type]

	nested := r.Rules
	if nested != nil && nested.Filter(filter).Size() > 0 {
		name = resourceTypeName(nested.ResourceType().String())
		if r.Type == Slice {
			name = "[]" + name
		}
	}

	return name
}

// resourceTypeName returns the human-readable type name for a resource.
func resourceTypeName(qualifiedName string) string {
	i := strings.LastIndex(qualifiedName, ".")
	if i < 0 {
		return qualifiedName
	}

	return qualifiedName[i+1 : len(qualifiedName)]
}

// handlerTypeName returns the human-readable type name for a ResourceHandler resource.
func handlerTypeName(handler ResourceHandler) string {
	rulesType := handler.Rules().ResourceType()
	if rulesType == nil {
		return handler.ResourceName()
	}

	return resourceTypeName(rulesType.String())
}

// fileName returns the constructed HTML file name for the provided name and version.
func fileName(name, version string) string {
	return strings.ToLower(fmt.Sprintf("%s_v%s.html", fileNamePrefix(name), version))
}

// fileNamePrefix returns the provided name as lower case with spaces replaced with
// underscores.
func fileNamePrefix(name string) string {
	return strings.ToLower(strings.Replace(name, " ", "_", -1))
}

// buildExampleRequest returns a JSON string representing an example endpoint request.
func buildExampleRequest(rules Rules, list bool, version string) string {
	return buildExamplePayload(rules, Inbound, list, version)
}

// buildExampleRequest returns a JSON string representing an example endpoint response.
func buildExampleResponse(rules Rules, list bool, version string) string {
	return buildExamplePayload(rules, Outbound, list, version)
}

// buildExamplePayload returns a JSON string representing either an example endpoint request
// or response depending on the Filter provided.
func buildExamplePayload(rules Rules, filter Filter, list bool, version string) string {
	rules = rules.ForVersion(version).Filter(filter)
	if rules.Size() == 0 {
		return ""
	}

	data := map[string]interface{}{}
	for _, r := range rules.Contents() {
		data[r.Name()] = getExampleValue(r, version)
	}

	var payload interface{}
	payload = data
	if list {
		payload = []interface{}{data}
	}

	serialized, err := json.MarshalIndent(payload, "", "    ")
	if err != nil {
		return ""
	}

	return string(serialized)
}

// getExampleValue returns an example value for the provided Rule.
func getExampleValue(r *Rule, version string) interface{} {
	value := r.DocExample
	if value != nil {
		return value
	}

	switch r.Type {
	case Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64:
		return 0
	case Float32, Float64:
		return 31.5
	case String:
		return "foo"
	case Bool:
		return true
	case Duration:
		return time.Duration(10000)
	case Time:
		return time.Date(2014, 9, 5, 15, 45, 36, 0, time.UTC)
	default:
		return getNestedExampleValue(r, version)
	}
}

// getNestedExampleValue returns an example value for a nested Rule value.
func getNestedExampleValue(r *Rule, version string) interface{} {
	if r.Rules == nil {
		switch r.Type {
		case Map:
			return map[string]interface{}{}
		case Slice:
			return []interface{}{}
		default:
			return nil
		}
	}

	ptr := reflect.New(r.Rules.ResourceType())
	value := applyOutboundRules(ptr.Elem().Interface(), r.Rules, version)
	if r.Type == Slice {
		value = []interface{}{value}
	}
	return value
}

// versions returns a slice containing all versions specified by the provided
// ResourceHandlers.
func versions(handlers []ResourceHandler) []string {
	versionMap := map[string]bool{}
	for _, handler := range handlers {
		for _, rule := range handler.Rules().Contents() {
			for _, version := range rule.Versions {
				versionMap[version] = true
			}
		}
	}

	versions := make([]string, 0, len(versionMap))
	for version := range versionMap {
		versions = append(versions, version)
	}

	sort.Strings(versions)
	return versions
}

// handlerVersions returns a slice containing all versions specified by the provided
// ResourceHandler.
func handlerVersions(handler ResourceHandler) []string {
	return versions([]ResourceHandler{handler})
}
