package rest

import (
	"log"
	"reflect"
)

// Rule provides fine-grained control over response output.
// TODO: Currently only supporting outbound Rules. Add support for inbound ones  which
// coerce types.
type Rule struct {
	// Name of the resource field.
	Field string

	// Name of the input/output field. Defaults to resource field name if not specified.
	ValueName string

	// Indicates if the Rule should only be applied to requests.
	InputOnly bool

	// Indicates if the Rule should only be applied to responses.
	OutputOnly bool

	// Function which produces the field value to send.
	Handler func(interface{}) interface{}
}

// applyOutboundRules applies Rules which are not specified as input only to the provided
// Resource. If the Resource is nil, not a struct, or no Rules are provided, this acts as
// an identity function. If Rules are provided, only the fields specified by them will be
// included in the returned Resource. This is to prevent new fields from leaking into old
// API versions.
func applyOutboundRules(resource Resource, rules []Rule) Resource {
	if resource == nil || len(rules) == 0 {
		// Return resource as-is if no Rules are provided.
		return resource
	}

	// Get the underlying value by dereferencing the pointer if there is one.
	resourceValue := reflect.Indirect(reflect.ValueOf(resource))
	resource = resourceValue.Interface()
	resourceType := reflect.TypeOf(resource)

	if resourceType.Kind() != reflect.Struct {
		// Only apply Rules to structs.
		// TODO: Can probably apply them to maps as well.
		return resource
	}

	payload := Payload{}

	for _, rule := range rules {
		if rule.InputOnly {
			// Apply only outbound Rules.
			continue
		}

		field := resourceValue.FieldByName(rule.Field)
		if !field.IsValid() {
			// The field doesn't exist.
			log.Printf("%s has no field %s", reflect.TypeOf(resource).Name(), rule.Field)
			continue
		}

		valueName := rule.ValueName
		if valueName == "" {
			// Use field name if value name isn't specified.
			valueName = rule.Field
		}

		fieldValue := field.Interface()
		if rule.Handler != nil {
			fieldValue = rule.Handler(fieldValue)
		}
		payload[valueName] = fieldValue
	}

	return payload
}
