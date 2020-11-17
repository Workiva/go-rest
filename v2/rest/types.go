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
	"reflect"
	"strconv"
	"time"
)

// Type is a data type to coerce a value to specified with a Rule.
type Type uint

// Type constants define the data types that Rules can specify for coercion.
const (
	Interface Type = iota
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Float32
	Float64
	String
	Bool
	Slice
	Map
	Duration
	Time
	Byte        = Uint8
	Unspecified = Interface
)

// typeToName maps Types to their human-readable names.
var typeToName = map[Type]string{
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
	Slice:     "[]interface{}",
	Map:       "map[string]interface{}",
	Duration:  "time.Duration",
	Time:      "time.Time",
}

// typeToKind maps Types to their reflect Kind.
var typeToKind = map[Type]reflect.Kind{
	Interface: reflect.Interface,
	Int:       reflect.Int,
	Int8:      reflect.Int8,
	Int16:     reflect.Int16,
	Int32:     reflect.Int32,
	Int64:     reflect.Int64,
	Uint:      reflect.Uint,
	Uint8:     reflect.Uint8,
	Uint16:    reflect.Uint16,
	Uint32:    reflect.Uint32,
	Uint64:    reflect.Uint64,
	Float32:   reflect.Float32,
	Float64:   reflect.Float64,
	String:    reflect.String,
	Bool:      reflect.Bool,
	Slice:     reflect.Slice,
	Map:       reflect.Map,
	Duration:  reflect.Int64,
	Time:      reflect.Struct,
}

// timeLayout is the format in which strings are parsed as time.Time (ISO 8601).
const timeLayout = "2006-01-02T15:04:05Z"

// coerceType attempts to convert the given value to the specified Type. If it cannot
// be coerced, nil will be returned along with an error.
func coerceType(value interface{}, coerceTo Type) (interface{}, error) {
	if coerceTo == Interface {
		return value, nil
	}

	// json.Unmarshal converts values to bool, float64, string, nil, slice, and map.
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
		return coerceFromSlice(value.([]interface{}), coerceTo)
	case map[string]interface{}:
		return coerceFromMap(value.(map[string]interface{}), coerceTo)
	default:
		return nil, fmt.Errorf("Unable to coerce %s to %s",
			reflect.TypeOf(value), typeToName[coerceTo])
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

	return nil, fmt.Errorf("Unable to coerce bool to %s", typeToName[coerceTo])
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

	// To Duration.
	case Duration:
		return time.Duration(int64(value)), nil

	// Bool and Time cases left off intentionally.
	default:
		return nil, fmt.Errorf("Unable to coerce float to %s", typeToName[coerceTo])
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

	// To Duration.
	case Duration:
		val, err := time.ParseDuration(value)
		if err != nil {
			return nil, err
		}
		return val, nil

	// To Time.
	case Time:
		val, err := time.Parse(timeLayout, value)
		if err != nil {
			return nil, err
		}
		return val, nil

	default:
		return nil, fmt.Errorf("Unable to coerce string to %s", typeToName[coerceTo])
	}
}

// coerceFromSlice attempts to convert the given slice to the specified Type. Currently,
// slices can only be coerced to slices (identity). If it cannot be coerced, nil will be
// returned along with an error.
func coerceFromSlice(value []interface{}, coerceTo Type) (interface{}, error) {
	if coerceTo == Slice {
		return value, nil
	}

	return nil, fmt.Errorf("Unable to coerce slice to %s", typeToName[coerceTo])
}

// coerceFromMap attempts to convert the given map to the specified Type. Currently,
// maps can only be coerced to maps (identity). If it cannot be coerced, nil will be
// returned along with an error.
func coerceFromMap(value map[string]interface{}, coerceTo Type) (interface{}, error) {
	if coerceTo == Map {
		return value, nil
	}

	return nil, fmt.Errorf("Unable to coerce map to %s", typeToName[coerceTo])
}
