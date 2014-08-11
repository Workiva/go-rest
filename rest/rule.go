package rest

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

// Type is a data type to coerce a value to specified with a Rule.
type Type uint

const (
	// Interface represents the interface{} data type.
	Interface Type = iota

	// Int represents the int data type.
	Int

	// Int8 represents the int8 data type.
	Int8

	// Int16 represents the int16 data type.
	Int16

	// Int32 represents the int32 data type.
	Int32

	// Int64 represents the int64 data type.
	Int64

	// Uint represents the uint data type.
	Uint

	// Uint8 represents the uint8 data type.
	Uint8

	// Uint16 represents the uint16 data type.
	Uint16

	// Uint32 represents the uint32 data type.
	Uint32

	// Uint64 represents the uint64 data type.
	Uint64

	// Float32 represents the float32 data type.
	Float32

	// Float64 represents the float64 data type.
	Float64

	// String represents the string data type.
	String

	// Bool represents the bool data type.
	Bool

	// Array represents the []interface{} data type.
	Array

	// Map represents the map[string]interface{} data type.
	Map

	// Byte represents the byte data type.
	Byte = Uint8

	// Unspecified represents the interface{} data type.
	Unspecified = Interface
)

// typeName maps Types to their human-readable names.
var typeName = map[Type]string{
	Interface: "interface{}",
	Int:       "int",
	Int8:      "int8",
	Int16:     "int16",
	Int32:     "int32",
	Int64:     "int64",
	Uint:      "uint",
	Uint8:     "uint8",
	Uint16:    "uint16",
	Uint32:    "uint32",
	Uint64:    "uint64",
	Float32:   "float32",
	Float64:   "float64",
	String:    "string",
	Bool:      "bool",
	Array:     "[]interface{}",
	Map:       "map[string]interface{}",
}

// Rule provides schema validation and type coercion for request input and fine-grained
// control over response output. If a ResourceHandler provides input Rules which specify
// types, input fields will attempt to be coerced to those types. If coercion fails, an
// error will be returned in the response. If a ResourceHandler provides output Rules,
// only the fields corresponding to those Rules will be sent back. This prevents new
// fields from leaking into old API versions.
type Rule struct {
	// Name of the resource field. This is the name as it appears in the struct
	// definition.
	Field string

	// Name of the input/output field. Defaults to resource field name if not specified.
	ValueName string

	// Type to coerce field value to. If the value cannot be coerced, an error will be
	// returned in the response. Defaults to Unspecified, which is the equivalent of
	// an interface{} value.
	Type Type

	// Indicates if the Rule should only be applied to requests.
	InputOnly bool

	// Indicates if the Rule should only be applied to responses.
	OutputOnly bool

	// Function which produces the field value to receive.
	InputHandler func(interface{}) interface{}

	// Function which produces the field value to send.
	OutputHandler func(interface{}) interface{}
}

// applyInboundRules applies Rules which are not specified as output only to the provided
// Payload. If the Payload is nil, an empty Payload will be returned. If no Rules are
// provided, this acts as an identity function. If Rules are provided, any incoming
// fields which are not specified will be discarded. If Rules specify types, incoming
// values will attempted to be coerced. If coercion fails, an error will be returned.
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
			if rule.ValueName == field {
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

	return newPayload, nil
}

// applyOutboundRules applies Rules which are not specified as input only to the provided
// Resource. If the Resource is nil, not a struct, or no Rules are provided, this acts as
// an identity function. If Rules are provided, only the fields specified by them will be
// included in the returned Resource. This is to prevent new fields from leaking into old
// API versions.
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
		field := resourceValue.FieldByName(rule.Field)
		if !field.IsValid() {
			// The field doesn't exist.
			log.Printf("%s has no field '%s'", reflect.TypeOf(resource).Name(), rule.Field)
			continue
		}

		valueName := rule.ValueName
		if valueName == "" {
			// Use field name if value name isn't specified.
			valueName = rule.Field
		}

		fieldValue := field.Interface()
		if rule.OutputHandler != nil {
			fieldValue = rule.OutputHandler(fieldValue)
		}
		payload[valueName] = fieldValue
	}

	return payload
}

// filterRules filters the array of Rules based on the specified bool. True means to
// filter out outbound Rules such that the returned array contains only inbound Rules.
// False means to filter out inbound Rules such that the returned array contains only
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

// coerceType attempts to convert the given value to the specified Type. If it cannot
// be coerced, nil will be returned along with an error.
func coerceType(value interface{}, coerceTo Type) (interface{}, error) {
	if coerceTo == Interface {
		return value, nil
	}

	// json.Unmarshal converts values to bool, float64, string, nil, array, and map.
	switch value.(type) {
	case bool:
		return coerceFromBool(value.(bool), coerceTo)
	case float64:
		return coerceFromFloat(value.(float64), coerceTo)
	case string:
		return coerceFromString(value.(string), coerceTo)
	case nil:
		return value, nil
	case []interface{}:
		return coerceFromArray(value.([]interface{}), coerceTo)
	case map[string]interface{}:
		return coerceFromMap(value.(map[string]interface{}), coerceTo)
	default:
		return nil, fmt.Errorf("Unable to coerce %s to %s",
			reflect.TypeOf(value), typeName[coerceTo])
	}
}

// coerceFromBool attempts to convert the given bool to the specified Type. If it
// cannot be coerced, nil will be returned along with an error.
func coerceFromBool(value bool, coerceTo Type) (interface{}, error) {
	if coerceTo == Bool {
		return value, nil
	}
	if coerceTo == String {
		if value {
			return "true", nil
		}
		return "false", nil
	}

	return nil, fmt.Errorf("Unable to coerce bool to %s", typeName[coerceTo])
}

// coerceFromFloat attempts to convert the given float64 to the specified Type.
// If it cannot be coerced, nil will be returned along with an error.
func coerceFromFloat(value float64, coerceTo Type) (interface{}, error) {
	switch coerceTo {
	// To int.
	case Int:
		return int(value), nil
	case Int8:
		return int8(value), nil
	case Int16:
		return int16(value), nil
	case Int32:
		return int32(value), nil
	case Int64:
		return int64(value), nil

	// To unsigned int.
	case Uint:
		return uint(value), nil
	case Uint8:
		return uint8(value), nil
	case Uint16:
		return uint16(value), nil
	case Uint32:
		return uint32(value), nil
	case Uint64:
		return uint64(value), nil

	// To float.
	case Float32:
		return float32(value), nil
	case Float64:
		return value, nil

	// To string.
	case String:
		return strconv.FormatFloat(value, 'f', -1, 64), nil

	// Bool case left off intentionally.
	default:
		return nil, fmt.Errorf("Unable to coerce float to %s", typeName[coerceTo])
	}
}

// coerceFromString attempts to convert the given string to the specified Type. If
// it cannot be coerced, nil will be returned along with an error.
func coerceFromString(value string, coerceTo Type) (interface{}, error) {
	switch coerceTo {
	// To int.
	case Int:
		val, err := strconv.ParseInt(value, 0, 0)
		if err != nil {
			return nil, err
		}
		return int(val), nil
	case Int8:
		val, err := strconv.ParseInt(value, 0, 8)
		if err != nil {
			return nil, err
		}
		return int8(val), nil
	case Int16:
		val, err := strconv.ParseInt(value, 0, 16)
		if err != nil {
			return nil, err
		}
		return int16(val), nil
	case Int32:
		val, err := strconv.ParseInt(value, 0, 32)
		if err != nil {
			return nil, err
		}
		return int32(val), nil
	case Int64:
		val, err := strconv.ParseInt(value, 0, 64)
		if err != nil {
			return nil, err
		}
		return int64(val), nil

	// To unsigned int.
	case Uint:
		val, err := strconv.ParseUint(value, 0, 0)
		if err != nil {
			return nil, err
		}
		return uint(val), nil
	case Uint8:
		val, err := strconv.ParseUint(value, 0, 8)
		if err != nil {
			return nil, err
		}
		return uint8(val), nil
	case Uint16:
		val, err := strconv.ParseUint(value, 0, 16)
		if err != nil {
			return nil, err
		}
		return uint16(val), nil
	case Uint32:
		val, err := strconv.ParseUint(value, 0, 32)
		if err != nil {
			return nil, err
		}
		return uint32(val), nil
	case Uint64:
		val, err := strconv.ParseUint(value, 0, 64)
		if err != nil {
			return nil, err
		}
		return uint64(val), nil

	// To float.
	case Float32:
		val, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, err
		}
		return float32(val), nil
	case Float64:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return float64(val), nil

	// To string.
	case String:
		return value, nil

	// To bool.
	case Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return val, nil

	default:
		return nil, fmt.Errorf("Unable to coerce string to %s", typeName[coerceTo])
	}
}

// coerceFromArray attempts to convert the given array to the specified Type. Currently,
// arrays can only be coerced to arrays (identity). If it cannot be coerced, nil will be
// returned along with an error.
func coerceFromArray(value []interface{}, coerceTo Type) (interface{}, error) {
	if coerceTo == Array {
		return value, nil
	}

	return nil, fmt.Errorf("Unable to coerce array to %s", typeName[coerceTo])
}

// coerceFromMap attempts to convert the given map to the specified Type. Currently,
// maps can only be coerced to maps (identity). If it cannot be coerced, nil will be
// returned along with an error.
func coerceFromMap(value map[string]interface{}, coerceTo Type) (interface{}, error) {
	if coerceTo == Map {
		return value, nil
	}

	return nil, fmt.Errorf("Unable to coerce map to %s", typeName[coerceTo])
}
