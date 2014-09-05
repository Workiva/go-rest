package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/hoisie/mustache"
)

type endpoint map[string]interface{}
type field map[string]interface{}
type handlerDoc map[string]string

func GenerateDocs(api API) {
	handlers := api.ResourceHandlers()
	docs := map[string][]handlerDoc{}
	versions := versions(handlers)

	for _, version := range versions {
		versionDocs := make([]handlerDoc, 0, len(handlers))
		for _, handler := range handlers {
			if doc, err := generateHandlerDoc(handler, version); err == nil {
				versionDocs = append(versionDocs, doc)
			}
		}

		docs[version] = versionDocs
	}

	generateIndexDoc(docs, versions)
}

func generateIndexDoc(docs map[string][]handlerDoc, versions []string) {
	tpl, err := mustache.ParseFile(
		"/Users/tylertreat/Go/src/github.com/WebFilings/go-rest/rest/index.mustache")
	if err != nil {
		log.Println(err)
		return
	}

	for version, docList := range docs {
		rendered := tpl.Render(map[string]interface{}{
			"handlers": docList,
			"version":  version,
		})
		ioutil.WriteFile(fmt.Sprintf("v%s.html", version), []byte(rendered), 0644)
	}

}

func generateHandlerDoc(handler ResourceHandler, version string) (handlerDoc, error) {
	handler = resourceHandlerProxy{handler}
	tpl, err := mustache.ParseFile(
		"/Users/tylertreat/Go/src/github.com/WebFilings/go-rest/rest/template.mustache")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	inputFields := getInputFields(handler.Rules().ForVersion(version))
	outputFields := getOutputFields(handler.Rules().ForVersion(version))

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
		return nil, fmt.Errorf("No documented endpoints")
	}

	name := handlerTypeName(handler)
	context := map[string]interface{}{
		"resource":  name,
		"version":   version,
		"endpoints": endpoints,
	}
	rendered := tpl.Render(context)

	file := fileName(name)
	ioutil.WriteFile(file, []byte(rendered), 0644)

	doc := handlerDoc{"name": name, "file": file}
	return doc, nil
}

func formatURI(uri, version string) string {
	uri = strings.Replace(uri, "{version:[^/]+}", version, -1)

	r, err := regexp.Compile("{.*?}")
	if err != nil {
		panic(err)
	}

	for _, param := range r.FindAllString(uri, -1) {
		uri = replaceURIParam(uri, param)
	}

	return uri
}

func replaceURIParam(uri, param string) string {
	paramName := param[1 : len(param)-1]
	return strings.Replace(uri, param, ":"+paramName, -1)
}

func getInputFields(rules Rules) []field {
	rules = rules.Filter(Inbound)
	fields := make([]field, 0, rules.Size())

	for _, rule := range rules.Contents() {
		required := "no"
		if rule.Required {
			required = "yes"
		}

		field := field{
			"name":     rule.Name(),
			"required": required,
			"type":     ruleTypeName(rule, Inbound),
		}

		fields = append(fields, field)
	}

	return fields
}

func getOutputFields(rules Rules) []field {
	rules = rules.Filter(Outbound)
	fields := make([]field, 0, rules.Size())

	for _, rule := range rules.Contents() {
		field := field{
			"name": rule.Name(),
			"type": ruleTypeName(rule, Outbound),
		}

		fields = append(fields, field)
	}

	return fields
}

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

func resourceTypeName(qualifiedName string) string {
	i := strings.LastIndex(qualifiedName, ".")
	if i < 0 {
		return qualifiedName
	}

	return qualifiedName[i+1 : len(qualifiedName)]
}

func handlerTypeName(handler ResourceHandler) string {
	rulesType := handler.Rules().ResourceType()
	if rulesType == nil {
		return handler.ResourceName()
	}

	return resourceTypeName(rulesType.String())
}

func fileName(name string) string {
	name = strings.Replace(name, " ", "_", -1)
	return strings.ToLower(name + ".html")
}

func buildExampleRequest(rules Rules, list bool, version string) string {
	return buildExamplePayload(rules, Inbound, list, version)
}

func buildExampleResponse(rules Rules, list bool, version string) string {
	return buildExamplePayload(rules, Outbound, list, version)
}

func buildExamplePayload(rules Rules, filter Filter, list bool, version string) string {
	rules = rules.Filter(filter)
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
		log.Println(err)
		return ""
	}

	return string(serialized)
}

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

func versions(handlers []ResourceHandler) []string {
	// TODO
	return []string{"1"}
}
