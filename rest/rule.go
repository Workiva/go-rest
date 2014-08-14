package rest

import (
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

// validateRules verifies that the provided Rules for are valid for the given
// reflect.Type, meaning they specify fields that exist and correct types. If a Rule
// is invalid, this will panic.
func validateRules(resourceType reflect.Type, rules []Rule) {
	if resourceType.Kind() != reflect.Struct && resourceType.Kind() != reflect.Map {
		panic(fmt.Sprintf("Invalid resource type: must be struct or map, got %s", resourceType))
	}

	for _, rule := range rules {
		field, found := resourceType.FieldByName(rule.Field)
		if !found {
			panic(fmt.Sprintf("Invalid Rule for %s: field '%s' does not exist",
				resourceType, rule.Field))
		}

		if !rule.validType(field.Type) {
			panic(fmt.Sprintf("Invalid Rule for %s: field '%s' is type %s, not %s",
				resourceType, rule.Field, field.Type, typeToName[rule.Type]))
		}
	}
}

// applyInboundRules applies Rules which are not specified as output only to the
// provided Payload. If the Payload is nil, an empty Payload will be returned. If no
// Rules are provided, this acts as an identity function. If Rules are provided, any
// incoming fields which are not specified will be discarded. If Rules specify types,
// incoming values will attempted to be coerced. If coercion fails, an error will be
// returned.
func applyInboundRules(payload Payload, rules []Rule) (Payload, error) {
	if payload == nil {
		return Payload{}, nil
	}

	// Apply only inbound Rules.
	rules = filterRules(rules, true)

	if len(rules) == 0 {
		return payload, nil
	}

	newPayload := Payload{}

fieldLoop:
	for field, value := range payload {
		for _, rule := range rules {
			if rule.FieldAlias == field {
				if rule.Type != Unspecified {
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

// applyOutboundRules applies Rules which are not specified as input only to the
// provided Resource. If the Resource is nil, not a struct, or no Rules are provided,
// this acts as an identity function. If Rules are provided, only the fields specified
// by them will be included in the returned Resource. This is to prevent new fields
// from leaking into old API versions.
func applyOutboundRules(resource Resource, rules []Rule) Resource {
	// Apply only outbound Rules.
	rules = filterRules(rules, false)

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
		// Rule validation occurs at server start. No need to check for field existence.
		field := resourceValue.FieldByName(rule.Field)
		fieldValue := field.Interface()
		if rule.OutputHandler != nil {
			fieldValue = rule.OutputHandler(fieldValue)
		}
		payload[rule.Name()] = fieldValue
	}

	return payload
}

// filterRules filters the slice of Rules based on the specified bool. True means to
// filter out outbound Rules such that the returned slice contains only inbound Rules.
// False means to filter out inbound Rules such that the returned slice contains only
// outbound Rules.
func filterRules(rules []Rule, inbound bool) []Rule {
	filtered := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		if inbound && rule.OutputOnly {
			// Filter out outbound Rules.
			continue
		} else if !inbound && rule.InputOnly {
			// Filter out inbound Rules.
			continue
		}
		filtered = append(filtered, rule)
	}

	return filtered
}

// enforceRequiredFields verifies that the provided Payload has values for any Rules
// with the Required flag set to true. If any required fields are missing, an error
// will be returned. Otherwise nil is returned.
func enforceRequiredFields(rules []Rule, payload Payload) error {
ruleLoop:
	for _, rule := range rules {
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
