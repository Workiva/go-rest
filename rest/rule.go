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
	"errors"
	"fmt"
	"log"
	"reflect"
)

// TODO:
//  - Consider using/writing a more robust data validation library.
//    E.g. https://github.com/Assembli/beautiful-validity
//    This would allow for semantic validation and custom validation logic.
//    For now, we are only providing type validation.
//
//  - Make type coercion pluggable (i.e. conversion/validation of custom types).

// Filter is a category for filtering Rules.
type Filter bool

// Rule category Filters.
const (
	Inbound  Filter = true
	Outbound Filter = false
)

// Rules is a collection of Rules and a reflect.Type which they correspond to.
type Rules interface {
	// Contents returns the contained Rules.
	Contents() []*Rule

	// ResourceType returns the reflect.Type these Rules correspond to.
	ResourceType() reflect.Type

	// Validate verifies that the Rules are valid, meaning they specify fields that exist
	// and correct types. If a Rule is invalid, an error is returned. If the Rules are
	// valid, nil is returned. This will recursively validate nested Rules.
	Validate() error

	// Filter will filter the Rules based on the specified Filter. Only Rules of the
	// specified Filter type will be returned.
	Filter(Filter) Rules

	// Size returns the number of contained Rules.
	Size() int

	// ForVersion returns the Rules which apply to the given version.
	ForVersion(string) Rules
}

type rules struct {
	contents     []*Rule
	resourceType reflect.Type
}

// Contents returns the contained Rules.
func (r *rules) Contents() []*Rule {
	return r.contents
}

// ResourceType returns the reflect.Type these Rules correspond to.
func (r *rules) ResourceType() reflect.Type {
	return r.resourceType
}

// Validate verifies that the Rules are valid, meaning they specify fields that exist
// and correct types. If a Rule is invalid, an error is returned. If the Rules are
// valid, nil is returned. This will recursively validate nested Rules.
func (r *rules) Validate() error {
	resourceType := r.resourceType
	if resourceType.Kind() != reflect.Struct && resourceType.Kind() != reflect.Map {
		return fmt.Errorf(
			"Invalid resource type: must be struct or map, got %s",
			resourceType)
	}

	for _, rule := range r.contents {
		if rule.Name() == "" {
			return fmt.Errorf("Invalid Rule: must have Field or FieldAlias")
		}

		if rule.isResourceRule() {
			if field, ok := resourceType.FieldByName(rule.Field); !ok {
				return fmt.Errorf(
					"Invalid Rule for %s: field '%s' does not exist",
					resourceType, rule.Field)
			} else if !rule.validType(field.Type) {
				return fmt.Errorf(
					"Invalid Rule for %s: field '%s' is type %s, not %s",
					resourceType, rule.Field, field.Type, typeToName[rule.Type])
			}
		}

		// Validate nested Rules.
		if rule.Rules != nil {
			// If a rule is on a slice, check to see what the underlying type is.
			// If it is primitive, there is nothing to validate.
			if typeToKind[rule.Type] == reflect.Slice {
				nestedType := typeToKind[rule.Rules.Contents()[0].Type]
				if nestedType == reflect.Struct || nestedType == reflect.Map {
					if err := rule.Rules.Validate(); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// Filter will filter the Rules based on the specified Filter. Only Rules of the
// specified Filter type will be returned.
func (r *rules) Filter(filter Filter) Rules {
	filtered := make([]*Rule, 0, len(r.contents))
	for _, rule := range r.contents {
		if filter == Inbound && rule.OutputOnly {
			// Filter out outbound Rules.
			continue
		} else if filter == Outbound && rule.InputOnly ||
			filter == Outbound && !rule.isResourceRule() {
			// Filter out inbound Rules.
			continue
		}
		filtered = append(filtered, rule)
	}

	return &rules{contents: filtered, resourceType: r.resourceType}
}

// Size returns the number of contained Rules.
func (r rules) Size() int {
	return len(r.contents)
}

// ForVersion returns the Rules which apply to the given version.
func (r *rules) ForVersion(version string) Rules {
	filtered := make([]*Rule, 0, r.Size())
	for _, rule := range r.Contents() {
		if rule.Applies(version) {
			filtered = append(filtered, rule)
		}
	}

	return &rules{contents: filtered, resourceType: r.resourceType}
}

// NewRules returns a set of Rules for use by a ResourceHandler. The first argument
// must be a resource pointer (and can be nil) used to associate the Rules with a
// resource type. If it isn't a pointer, this will panic.
func NewRules(ptr interface{}, r ...*Rule) Rules {
	resourceType := reflect.TypeOf(ptr)
	if resourceType.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Must provide resource pointer to NewRules, got %s",
			resourceType.Kind()))
	}

	return &rules{
		resourceType: resourceType.Elem(),
		contents:     r,
	}
}

// Rule provides schema validation and type coercion for request input and fine-grained
// control over response output. If a ResourceHandler provides input Rules which
// specify types, input fields will attempt to be coerced to those types. If coercion
// fails, an error will be returned in the response. If a ResourceHandler provides
// output Rules, only the fields corresponding to those Rules will be sent back. This
// prevents new fields from leaking into old API versions.
type Rule struct {
	// Name of the resource field. This is the name as it appears in the struct
	// definition.
	Field string

	// Name of the input/output field. Use Name() to retrieve the field alias while
	// falling back to the field name if it's not specified.
	FieldAlias string

	// Type to coerce field value to. If the value cannot be coerced, an error will be
	// returned in the response. Defaults to Unspecified, which is the equivalent of
	// an interface{} value.
	Type Type

	// Indicates if the field must have a value. Defaults to false.
	Required bool

	// Versions is a list of the API versions this Rule applies to. If empty, it will
	// be applied to all versions.
	Versions []string

	// Indicates if the Rule should only be applied to requests.
	InputOnly bool

	// Indicates if the Rule should only be applied to responses.
	OutputOnly bool

	// Function which produces the field value to receive.
	InputHandler func(interface{}) interface{}

	// Function which produces the field value to send.
	OutputHandler func(interface{}) interface{}

	// Nested Rules to apply to field value.
	Rules Rules

	// Description used in documentation.
	DocString string

	// Example value used in documentation.
	DocExample interface{}
}

// Name returns the name of the input/output field alias. It defaults to the field
// name if the alias was not specified.
func (r Rule) Name() string {
	alias := r.FieldAlias
	if alias == "" {
		alias = r.Field
	}
	return alias
}

// Applies returns whether or not the Rule applies to the given version.
func (r Rule) Applies(version string) bool {
	if r.Versions == nil {
		return true
	}

	for _, v := range r.Versions {
		if v == version {
			return true
		}
	}

	return false
}

// validType returns whether or not the Rule is valid for the given reflect.Type.
func (r Rule) validType(fieldType reflect.Type) bool {
	if r.Type == Unspecified {
		return true
	}

	kind := typeToKind[r.Type]
	return fieldType.Kind() == kind
}

// isResourceRule returns true if this Rule corresponds to a resource field, false
// if not. Non-resource Rules allow you to specify input fields that do not directly
// correspond to a resource.
func (r Rule) isResourceRule() bool {
	return r.Field != ""
}

// applyInboundRules applies Rules which are not specified as output only to the
// provided Payload. If the Payload is nil, an empty Payload will be returned. If no
// Rules are provided, this acts as an identity function. If Rules are provided, any
// incoming fields which are not specified will be discarded. If Rules specify types,
// incoming values will attempted to be coerced. If coercion fails, an error will be
// returned. If Rules specify nested Rules, they will be recursively applied to the
// field value, taking precedence over a type coercion.
func applyInboundRules(payload Payload, rules Rules, version string) (Payload, error) {
	if payload == nil {
		return Payload{}, nil
	}

	// Apply only inbound Rules.
	rules = rules.Filter(true).ForVersion(version)

	if rules.Size() == 0 {
		return payload, nil
	}

	newPayload := Payload{}

fieldLoop:
	for field, value := range payload {
		for _, rule := range rules.Contents() {
			if rule.Name() == field {
				if nestedInboundRulesApply(value, rule.Rules, version) {
					// Nested Rules take precedence over type coercion.
					v, err := applyNestedInboundRules(value, rule.Rules, version)
					if err != nil {
						return nil, err
					}
					value = v
				} else if rule.Type != Unspecified {
					// Coerce to specified type.
					coerced, err := coerceType(value, rule.Type)
					if err != nil {
						return nil, err
					}
					value = coerced
				}

				if rule.InputHandler != nil {
					value = rule.InputHandler(value)
				}

				newPayload[field] = value
				continue fieldLoop
			}
		}

		log.Printf("Discarding field '%s'", field)
	}

	// Ensure no required fields are missing.
	if err := enforceRequiredFields(rules, newPayload); err != nil {
		log.Println(err)
		return nil, err
	}

	return newPayload, nil
}

// applyNestedInboundRules recursively applies nested Rules which are not specified as
// output only to the provided value.
func applyNestedInboundRules(
	value interface{}, rules Rules, version string) (interface{}, error) {

	var fieldValue interface{}
	valueType := reflect.TypeOf(value).Kind()
	if valueType == reflect.Slice {
		// Apply nested Rules to each item in the slice.
		s := reflect.ValueOf(value)
		nestedValues := make([]interface{}, s.Len())
		for i := 0; i < s.Len(); i++ {
			val := s.Index(i).Interface()
			if val == nil {
				return nil, errors.New("nested value is nil")
			}
			// Check to see if the nested type is a slice or map.
			// If not, it should be coerced to its rule type.
			iKind := reflect.TypeOf(val).Kind()
			ruleType := rules.Contents()[0].Type
			if ruleKind := typeToKind[ruleType]; iKind == reflect.Slice && ruleKind != reflect.Slice {
				return nil, fmt.Errorf("Value does not match rule type, expecting: %v, got: %v", ruleKind, iKind)
			} else if iKind != reflect.Map && iKind != reflect.Slice {
				payloadIFace, err := coerceType(s.Index(i).Interface(), ruleType)
				if err != nil {
					return nil, err
				}
				nestedValues[i] = payloadIFace
			} else {
				payloadIFace, err := coerceType(s.Index(i).Interface(), Map)
				if err != nil {
					return nil, err
				}
				var payload map[string]interface{}
				payload, err = applyInboundRules(payloadIFace.(map[string]interface{}), rules, version)
				if err != nil {
					return nil, err
				}
				nestedValues[i] = payload
			}
		}
		fieldValue = nestedValues
	} else {
		payloadIFace, err := coerceType(value, Map)
		if err != nil {
			return nil, err
		}
		var payload map[string]interface{}
		payload, err = applyInboundRules(payloadIFace.(map[string]interface{}), rules, version)
		if err != nil {
			return nil, err
		}
		fieldValue = payload
	}

	return fieldValue, nil
}

// nestedInboundRulesApply returns true if the Rules contain inbound Rules and
// the value is a map or slice.
func nestedInboundRulesApply(value interface{}, rules Rules, version string) bool {
	if rules == nil || rules.Size() == 0 {
		return false
	}

	valueType := reflect.TypeOf(value).Kind()
	if valueType != reflect.Map && valueType != reflect.Slice {
		// Only apply nested Rules to maps and slices.
		return false
	}

	return rules.Filter(Inbound).ForVersion(version).Size() > 0
}

// applyOutboundRules applies Rules which are not specified as input only to the
// provided Resource. If the Resource is nil, not a struct or
// map[string]interface{}, or no Rules are provided, this acts as an identity
// function. If Rules are provided, only the fields specified by them will be
// included in the returned Resource. This is to prevent new fields from leaking
// into old API versions. If Rules specify nested Rules, they will be recursively
// applied to field values.
func applyOutboundRules(resource Resource, rules Rules, version string) Resource {
	// Apply only outbound Rules.
	rules = rules.Filter(false).ForVersion(version)

	if isNil(resource) || rules.Size() == 0 {
		// Return resource as-is if no Rules are provided.
		return resource
	}

	// Get the underlying value by dereferencing the pointer if there is one.
	resourceValue := reflect.Indirect(reflect.ValueOf(resource))
	resource = resourceValue.Interface()
	resourceType := reflect.TypeOf(resource)
	var payload Resource

	if resourceType.Kind() == reflect.Map {
		if resourceMap, ok := resource.(map[string]interface{}); ok {
			payload = applyOutboundRulesForMap(resourceMap, rules, version)
		} else {
			// Nothing we can do if the keys aren't strings.
			payload = resource
		}
	} else if resourceType.Kind() == reflect.Struct {
		payload = applyOutboundRulesForStruct(resourceValue, rules, version)
	} else {
		// Only apply Rules to resource structs and maps.
		payload = resource
	}

	return payload
}

// applyOutboundRulesForMap applies Rules which are not specified as input only to the
// provided map. If a Rule specifies a field which is not in the map, it will be skipped.
// If a Rule specifies nested Rules, they will be recursively applied to the corresponding
// value.
func applyOutboundRulesForMap(
	resource map[string]interface{}, rules Rules, version string) Payload {

	payload := Payload{}
	for _, rule := range rules.Contents() {
		if !rule.isResourceRule() {
			// Non-resource Rules don't apply to output.
			continue
		}

		fieldValue, ok := resource[rule.Field]
		if !ok {
			log.Printf("Map resource missing field '%s'", rule.Field)
			continue
		}

		if rule.Rules != nil {
			fieldValue = applyNestedOutboundRules(fieldValue, rule, version)
		}

		if rule.OutputHandler != nil {
			fieldValue = rule.OutputHandler(fieldValue)
		}
		payload[rule.Name()] = fieldValue
	}

	return payload
}

// applyOutboundRulesForStruct applies Rules which are not specified as input only to the
// provided reflect.Value. The precondition for this function is that the value is an
// instance of the type specified on the Rules. If a Rule specifies nested Rules, they
// will be recursively applied to the corresponding value.
func applyOutboundRulesForStruct(
	resourceValue reflect.Value, rules Rules, version string) Payload {

	payload := Payload{}
	for _, rule := range rules.Contents() {
		if !rule.isResourceRule() {
			// Non-resource Rules don't apply to output.
			continue
		}

		// Rule validation occurs at server start. No need to check for field existence.
		field := resourceValue.FieldByName(rule.Field)
		fieldValue := field.Interface()

		if rule.Rules != nil {
			fieldValue = applyNestedOutboundRules(fieldValue, rule, version)
		}

		if rule.OutputHandler != nil {
			fieldValue = rule.OutputHandler(fieldValue)
		}
		payload[rule.Name()] = fieldValue
	}

	return payload
}

// applyNestedOutboundRules recursively applies nested Rules which are not specified as
// input only to the provided Resource.
func applyNestedOutboundRules(resource Resource, rule *Rule, version string) Resource {
	var fieldValue Resource

	if reflect.TypeOf(resource).Kind() == reflect.Slice {
		// Apply nested Rules to each item in the slice.
		s := reflect.ValueOf(resource)
		nestedValues := make([]interface{}, s.Len())
		for i := 0; i < s.Len(); i++ {
			nestedValues[i] = applyOutboundRules(
				s.Index(i).Interface(), rule.Rules, version)
		}
		fieldValue = nestedValues
	} else {
		fieldValue = applyOutboundRules(resource, rule.Rules, version)
	}

	return fieldValue
}

// enforceRequiredFields verifies that the provided Payload has values for any Rules
// with the Required flag set to true. If any required fields are missing, an error
// will be returned. Otherwise nil is returned.
func enforceRequiredFields(rules Rules, payload Payload) error {
ruleLoop:
	for _, rule := range rules.Contents() {
		if !rule.Required {
			continue
		}

		for field := range payload {
			if rule.Name() == field {
				continue ruleLoop
			}
		}

		return fmt.Errorf("Missing required field '%s'", rule.Name())
	}

	return nil
}

// isNil returns true if the given Resource is a nil value or pointer, false if
// not.
func isNil(resource Resource) bool {
	value := reflect.ValueOf(resource)
	t := value.Kind()
	switch t {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return resource == nil
	}
}
