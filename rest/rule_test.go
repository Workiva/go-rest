package rest

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Ensures that an empty Payload is returned by applyInboundRules if nil is passed in.
func TestApplyInboundRulesNilPayload(t *testing.T) {
	assert := assert.New(t)

	actual, err := applyInboundRules(nil, NewRules((*TestResource)(nil)), "1")

	assert.Equal(Payload{}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that payload is returned by applyInboundRules if rules is empty.
func TestApplyInboundRulesNoRules(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{}

	actual, err := applyInboundRules(payload, NewRules((*TestResource)(nil)), "1")

	assert.Equal(payload, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that an error is returned by applyInboundRules if a required field is
// missing.
func TestApplyInboundRulesMissingRequiredField(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := applyInboundRules(payload, NewRules((*TestResource)(nil),
		&Rule{
			Field:    "baz",
			Required: true,
		},
		&Rule{
			Field:    "foo",
			Required: true,
		},
	), "1")

	assert.Nil(actual, "Return value should be nil")
	assert.Equal(fmt.Errorf("Missing required field 'baz'"), err, "Incorrect error")
}

// Ensures that only inbound rules are applied and unspecified input fields are discarded.
func TestApplyInboundRules(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar", "baz": 1}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": "bar"}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify an input handler yield the correct values.
func TestApplyInboundRulesInputHandler(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar", "baz": 1}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:        "foo",
			FieldAlias:   "foo",
			InputHandler: func(val interface{}) interface{} { return val.(string) + " qux" },
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": "bar qux"}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensure that if type coercion from bool fails, the error is returned.
func TestApplyInboundRulesCoerceBoolError(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": true}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Float32,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Nil(actual, "Return value should be nil")
	assert.Equal(fmt.Errorf("Unable to coerce bool to float32"), err, "Incorrect error")
}

// Ensures that inbound rules which specify bool correctly coerce bool.
func TestApplyInboundRulesCoerceBoolToBool(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": true}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Bool,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": true}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify string correctly coerce bool.
func TestApplyInboundRulesCoerceBoolToString(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": true}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       String,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": "true"}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensure that if type coercion from float64 fails, the error is returned.
func TestApplyInboundRulesCoerceFloatError(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Bool,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Nil(actual, "Return value should be nil")
	assert.Equal(fmt.Errorf("Unable to coerce float to bool"), err, "Incorrect error")
}

// Ensures that inbound rules which specify int correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToInt(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Int,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": int(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify int8 correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToInt8(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Int8,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": int8(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify int16 correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToInt16(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Int16,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": int16(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify int32 correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToInt32(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Int32,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": int32(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify int64 correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToInt64(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Int64,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": int64(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify uint correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToUint(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Uint,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": uint(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify uint8 correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToUint8(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Uint8,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": uint8(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify uint16 correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToUint16(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Uint16,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": uint16(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify uint32 correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToUint32(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Uint32,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": uint32(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify uint64 correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToUint64(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Uint64,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": uint64(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify float32 correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToFloat32(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Float32,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": float32(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify float64 correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToFloat64(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Float64,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": float64(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify string correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToString(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       String,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": "42"}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify time.Duration correctly coerce float64.
func TestApplyInboundRulesCoerceFloatToDuration(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(42)}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Duration,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": time.Duration(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensure that if type coercion from string fails, the error is returned.
func TestApplyInboundRulesCoerceStringError(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "hello"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Map,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Nil(actual, "Return value should be nil")
	assert.Equal(fmt.Errorf("Unable to coerce string to map[string]interface{}"),
		err, "Incorrect error")
}

// Ensure that if type coercion from string to int fails, the error is returned.
func TestApplyInboundRulesCoerceStringIntError(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "hello"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Int,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Nil(actual, "Return value should be nil")
	assert.Equal("strconv.ParseInt: parsing \"hello\": invalid syntax",
		err.Error(), "Incorrect error")
}

// Ensures that inbound rules which specify int correctly coerce string.
func TestApplyInboundRulesCoerceStringToInt(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Int,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": int(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify int8 correctly coerce string.
func TestApplyInboundRulesCoerceStringToInt8(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Int8,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": int8(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify int16 correctly coerce string.
func TestApplyInboundRulesCoerceStringToInt16(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Int16,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": int16(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify int32 correctly coerce string.
func TestApplyInboundRulesCoerceStringToInt32(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Int32,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": int32(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify int64 correctly coerce string.
func TestApplyInboundRulesCoerceStringToInt64(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Int64,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": int64(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensure that if type coercion from string to uint fails, the error is returned.
func TestApplyInboundRulesCoerceStringUintError(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "hello"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Uint,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Nil(actual, "Return value should be nil")
	assert.Equal("strconv.ParseUint: parsing \"hello\": invalid syntax",
		err.Error(), "Incorrect error")
}

// Ensures that inbound rules which specify uint correctly coerce string.
func TestApplyInboundRulesCoerceStringToUint(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Uint,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": uint(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify uint8 correctly coerce string.
func TestApplyInboundRulesCoerceStringToUint8(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Uint8,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": uint8(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify uint16 correctly coerce string.
func TestApplyInboundRulesCoerceStringToUint16(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Uint16,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": uint16(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify uint32 correctly coerce string.
func TestApplyInboundRulesCoerceStringToUint32(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Uint32,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": uint32(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify uint64 correctly coerce string.
func TestApplyInboundRulesCoerceStringToUint64(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Uint64,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": uint64(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensure that if type coercion from string to float fails, the error is returned.
func TestApplyInboundRulesCoerceStringFloatError(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "hello"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Float32,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Nil(actual, "Return value should be nil")
	assert.Equal("strconv.ParseFloat: parsing \"hello\": invalid syntax",
		err.Error(), "Incorrect error")
}

// Ensures that inbound rules which specify float32 correctly coerce string.
func TestApplyInboundRulesCoerceStringToFloat32(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Float32,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": float32(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify float64 correctly coerce string.
func TestApplyInboundRulesCoerceStringToFloat64(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Float64,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": float64(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify string correctly coerce string.
func TestApplyInboundRulesCoerceStringToString(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       String,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": "42"}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensure that if type coercion from string to bool fails, the error is returned.
func TestApplyInboundRulesCoerceStringBoolError(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "hello"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Bool,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Nil(actual, "Return value should be nil")
	assert.Equal("strconv.ParseBool: parsing \"hello\": invalid syntax",
		err.Error(), "Incorrect error")
}

// Ensures that inbound rules which specify bool correctly coerce string.
func TestApplyInboundRulesCoerceStringToBool(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "true"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Bool,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": true}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify time.Duration correctly coerce string.
func TestApplyInboundRulesCoerceStringToDurationError(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "hello"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Duration,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Nil(actual, "Return value should be nil")
	assert.Equal("time: invalid duration hello", err.Error(), "Incorrect error")
}

// Ensures that inbound rules which specify time.Duration correctly coerce string.
func TestApplyInboundRulesCoerceStringToDuration(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "100ns"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Duration,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": time.Duration(100)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that inbound rules which specify time.Time correctly coerce string.
func TestApplyInboundRulesCoerceStringToTimeError(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "hello"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Time,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Nil(actual, "Return value should be nil")
	assert.Equal(
		"parsing time \"hello\" as \"2006-01-02T15:04:05Z\": "+
			"cannot parse \"hello\" as \"2006\"", err.Error(), "Incorrect error")
}

// Ensures that inbound rules which specify time.Time correctly coerce string.
func TestApplyInboundRulesCoerceStringToTime(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "2014-08-11T15:46:02Z"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Time,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": time.Date(2014, 8, 11, 15, 46, 2, 0, time.UTC)},
		actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensure that if type coercion from slice fails, the error is returned.
func TestApplyInboundRulesCoerceSliceError(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": []interface{}{1, 2, 3}}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Bool,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Nil(actual, "Return value should be nil")
	assert.Equal(fmt.Errorf("Unable to coerce slice to bool"), err, "Incorrect error")
}

// Ensures that inbound rules which specify slice correctly coerce slice.
func TestApplyInboundRulesCoerceSliceToSlice(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": []interface{}{1, 2, 3}}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Slice,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": []interface{}{1, 2, 3}}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensure that if type coercion from map fails, the error is returned.
func TestApplyInboundRulesCoerceMapError(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": map[string]interface{}{"a": 1}}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Bool,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Nil(actual, "Return value should be nil")
	assert.Equal(fmt.Errorf("Unable to coerce map to bool"), err, "Incorrect error")
}

// Ensures that inbound rules which specify map correctly coerce map.
func TestApplyInboundRulesCoerceMapToMap(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": map[string]interface{}{"a": 1}}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "foo",
			FieldAlias: "foo",
			Type:       Map,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": map[string]interface{}{"a": 1}},
		actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that nested inbound Rules are not applied if the field value is not a map
// or slice.
func TestApplyInboundRulesNestedRulesDontApply(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": 1}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
			Rules: NewRules((*TestResource)(nil),
				&Rule{Field: "Bar", FieldAlias: "bar", Type: String}),
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": 1}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that nested inbound Rules are correctly applied to maps.
func TestApplyInboundRulesNestedRulesMap(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": map[string]interface{}{"bar": "baz"}}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
			Rules: NewRules((*TestResource)(nil),
				&Rule{Field: "Bar", FieldAlias: "bar", Type: String}),
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": map[string]interface{}{"bar": "baz"}}, actual,
		"Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that nested inbound Rules are correctly applied to slices.
func TestApplyInboundRulesNestedRulesSlice(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": []map[string]interface{}{
		map[string]interface{}{"bar": "a"},
		map[string]interface{}{"bar": "b"},
	}}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
			Rules: NewRules((*TestResource)(nil),
				&Rule{Field: "Bar", FieldAlias: "bar", Type: String}),
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": []interface{}{
		map[string]interface{}{"bar": "a"},
		map[string]interface{}{"bar": "b"},
	}}, actual, "Incorrect return value")

	assert.Nil(err, "Error should be nil")
}

// Ensures that nested inbound Rules are not applied if they do not apply to the version.
func TestApplyInboundRulesNestedRulesDifferentVersion(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": []map[string]interface{}{
		map[string]interface{}{"foo": "hi", "bar": "a"},
		map[string]interface{}{"foo": "hello", "bar": "b"},
	}}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
			Versions:   []string{"1"},
			Rules: NewRules((*TestResource)(nil),
				&Rule{
					FieldAlias: "foo",
					Versions:   []string{"1"},
				},
				&Rule{
					Field:      "Bar",
					FieldAlias: "bar",
					Type:       String,
					Versions:   []string{"2"},
				},
			),
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": []interface{}{
		map[string]interface{}{"foo": "hi"},
		map[string]interface{}{"foo": "hello"},
	}}, actual, "Incorrect return value")

	assert.Nil(err, "Error should be nil")
}

// Ensures that non-resource Rules are applied correctly.
func TestApplyInboundRulesNonResourceRule(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "42"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			FieldAlias: "foo",
			Type:       Float32,
		},
	)

	actual, err := applyInboundRules(payload, rules, "1")

	assert.Equal(Payload{"foo": float32(42)}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that nil is returned by applyOutboundRules if nil is passed in.
func TestApplyOutboundRulesNilResource(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(applyOutboundRules(
		nil, NewRules((*TestResource)(nil), &Rule{}), "1"), "Incorrect return value")
}

// Ensures that resource is returned by applyOutboundRules if rules is empty.
func TestApplyOutboundRulesNoRules(t *testing.T) {
	assert := assert.New(t)
	resource := &TestResource{}

	assert.Equal(resource, applyOutboundRules(
		resource, NewRules((*TestResource)(nil)), "1"), "Incorrect return value")
}

// Ensures that resource is returned by applyOutboundRules if it's not a struct.
func TestApplyOutboundRulesNotStruct(t *testing.T) {
	assert := assert.New(t)
	resource := "resource"

	assert.Equal(resource, applyOutboundRules(
		resource, NewRules((*TestResource)(nil), &Rule{}), "1"), "Incorrect return value")
}

// Ensures that applyOutboundRules handles map[string]interface.
func TestApplyOutboundRulesMap(t *testing.T) {
	assert := assert.New(t)
	resource := map[string]interface{}{"Foo": "hello"}
	rules := NewRules((*TestResource)(nil), &Rule{Field: "Foo"})

	assert.Equal(
		Payload{"Foo": "hello"},
		applyOutboundRules(resource, rules, "1"),
		"Incorrect return value",
	)
}

// Ensures that applyOutboundRules handles map[string]interface and missing fields
// are ignored.
func TestApplyOutboundRulesMapMissingFields(t *testing.T) {
	assert := assert.New(t)
	resource := map[string]interface{}{}
	rules := NewRules((*TestResource)(nil), &Rule{Field: "Foo"})

	assert.Equal(
		Payload{},
		applyOutboundRules(resource, rules, "1"),
		"Incorrect return value",
	)
}

// Ensures that rules which specify an output Handler function yield the correct value
// for maps.
func TestApplyOutboundRulesMapOutputHandler(t *testing.T) {
	assert := assert.New(t)
	resource := map[string]interface{}{"Foo": "hello"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
			OutputHandler: func(val interface{}) interface{} {
				return val.(string) + " world"
			},
		},
	)

	assert.Equal(
		Payload{"foo": "hello world"},
		applyOutboundRules(resource, rules, "1"),
		"Incorrect return value",
	)
}

// Ensures that resource is returned by applyOutboundRules if it's an incorrect map
// type.
func TestApplyOutboundRulesBadMap(t *testing.T) {
	assert := assert.New(t)
	resource := map[int]interface{}{}
	rules := NewRules((*TestResource)(nil), &Rule{Field: "Foo"})

	assert.Equal(
		resource,
		applyOutboundRules(resource, rules, "1"),
		"Incorrect return value",
	)
}

// Ensures that nested Rules are properly applied to slices by applyOutboundRules
// when a map is passed in.
func TestApplyOutboundRulesMapNestedRulesSlice(t *testing.T) {
	assert := assert.New(t)
	resource := map[string]interface{}{
		"Foo": "bar",
		"Baz": []TestResource{TestResource{Foo: "hello"}},
	}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
		},
		&Rule{
			Field:      "Baz",
			FieldAlias: "baz",
			Rules: NewRules((*TestResource)(nil),
				&Rule{
					Field:      "Foo",
					FieldAlias: "f",
				},
			),
		},
	)

	assert.Equal(
		Payload{"foo": "bar", "baz": []interface{}{Payload{"f": "hello"}}},
		applyOutboundRules(resource, rules, "1"),
		"Incorrect return value",
	)
}

// Ensures that nested Rules are properly applied to non-slice values by
// applyOutboundRules when a map is passed in.
func TestApplyOutboundRulesMapNestedRulesValueNonSlice(t *testing.T) {
	assert := assert.New(t)
	resource := map[string]interface{}{
		"Foo": "bar",
		"Baz": TestResource{Foo: "hello"},
	}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
		},
		&Rule{
			Field:      "Baz",
			FieldAlias: "baz",
			Rules: NewRules((*TestResource)(nil),
				&Rule{
					Field:      "Foo",
					FieldAlias: "f",
				},
			),
		},
	)

	assert.Equal(
		Payload{"foo": "bar", "baz": Payload{"f": "hello"}},
		applyOutboundRules(resource, rules, "1"),
		"Incorrect return value",
	)
}

// Ensures that nested Rules are properly applied to slices by applyOutboundRules
// when a struct is passed in.
func TestApplyOutboundRulesStructNestedRulesSlice(t *testing.T) {
	assert := assert.New(t)
	type nestedResource struct {
		Foo string
		Bar []TestResource
	}
	resource := nestedResource{
		Foo: "hello",
		Bar: []TestResource{TestResource{Foo: "world"}},
	}
	rules := NewRules((*nestedResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
		},
		&Rule{
			Field:      "Bar",
			FieldAlias: "bar",
			Rules: NewRules((*TestResource)(nil),
				&Rule{
					Field:      "Foo",
					FieldAlias: "f",
				},
			),
		},
	)

	assert.Equal(
		Payload{"foo": "hello", "bar": []interface{}{Payload{"f": "world"}}},
		applyOutboundRules(resource, rules, "1"),
		"Incorrect return value",
	)
}

// Ensures that nested Rules are properly applied to non-slice values by
// applyOutboundRules when a struct is passed in.
func TestApplyOutboundRulesStructNestedRulesValueNonSlice(t *testing.T) {
	assert := assert.New(t)
	type nestedResource struct {
		Foo string
		Bar TestResource
	}
	resource := nestedResource{
		Foo: "hello",
		Bar: TestResource{Foo: "world"},
	}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
		},
		&Rule{
			Field:      "Bar",
			FieldAlias: "bar",
			Rules: NewRules((*TestResource)(nil),
				&Rule{
					Field:      "Foo",
					FieldAlias: "f",
				},
			),
		},
	)

	assert.Equal(
		Payload{"foo": "hello", "bar": Payload{"f": "world"}},
		applyOutboundRules(resource, rules, "1"),
		"Incorrect return value",
	)
}

// Ensures that nested outbound Rules are not applied if the version doesn't apply.
func TestApplyOutboundRulesNestedRulesDifferentVersion(t *testing.T) {
	assert := assert.New(t)
	type nestedResource struct {
		Foo string
		Bar []TestResource
	}
	resource := nestedResource{
		Foo: "hello",
		Bar: []TestResource{TestResource{Foo: "world"}},
	}
	rules := NewRules((*nestedResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
			Versions:   []string{"1"},
		},
		&Rule{
			Field:      "Bar",
			FieldAlias: "bar",
			Versions:   []string{"1"},
			Rules: NewRules((*TestResource)(nil),
				&Rule{
					Field:      "Foo",
					FieldAlias: "f",
					Versions:   []string{"2"},
				},
			),
		},
	)

	assert.Equal(
		Payload{"foo": "hello", "bar": []interface{}{TestResource{Foo: "world"}}},
		applyOutboundRules(resource, rules, "1"),
		"Incorrect return value",
	)
}

// Ensures that rules which don't specify a valueName use the field name by default.
func TestApplyOutboundRulesDefaultName(t *testing.T) {
	assert := assert.New(t)
	resource := &TestResource{Foo: "hello"}
	rules := NewRules((*TestResource)(nil), &Rule{Field: "Foo"})

	assert.Equal(
		Payload{"Foo": "hello"},
		applyOutboundRules(resource, rules, "1"),
		"Incorrect return value",
	)
}

// Ensures that rules which specify an output Handler function yield the correct value.
func TestApplyOutboundRulesOutputHandler(t *testing.T) {
	assert := assert.New(t)
	resource := &TestResource{Foo: "hello"}
	rules := NewRules((*TestResource)(nil),
		&Rule{
			Field:      "Foo",
			FieldAlias: "foo",
			OutputHandler: func(val interface{}) interface{} {
				return val.(string) + " world"
			},
		},
	)

	assert.Equal(
		Payload{"foo": "hello world"},
		applyOutboundRules(resource, rules, "1"),
		"Incorrect return value",
	)
}

// Ensures that non-resource Rules are ignored on output.
func TestApplyOutboundRulesNonResourceRule(t *testing.T) {
	assert := assert.New(t)
	resource := &TestResource{Foo: "hello"}
	rules := NewRules((*TestResource)(nil), &Rule{FieldAlias: "bar"})

	assert.Equal(
		&TestResource{Foo: "hello"},
		applyOutboundRules(resource, rules, "1"),
		"Incorrect return value",
	)
}

// Ensures that Applies returns true if no versions are specified on the Rule.
func TestAppliesNoVersions(t *testing.T) {
	assert := assert.New(t)
	rule := &Rule{}

	assert.True(rule.Applies("v1"), "Incorrect return value")
}

// Ensures that Applies returns false if the version is not in the Rule's versions.
func TestAppliesDoesNotApply(t *testing.T) {
	assert := assert.New(t)
	rule := &Rule{Versions: []string{"v2"}}

	assert.False(rule.Applies("v1"), "Incorrect return value")
}

// Ensures that Applies returns true if the version is in the Rule's versions.
func TestAppliesDoesApply(t *testing.T) {
	assert := assert.New(t)
	rule := &Rule{Versions: []string{"v1", "v2", "v3"}}

	assert.True(rule.Applies("v2"), "Incorrect return value")
}

// Ensures that isNil returns true for nil value.
func TestIsNilNilValue(t *testing.T) {
	assert := assert.New(t)
	assert.True(isNil(nil), "Incorrect return value")
}

// Ensures that isNil returns true for nil pointer.
func TestIsNilNilPtr(t *testing.T) {
	assert := assert.New(t)
	var ptr *TestResource
	assert.True(isNil(ptr), "Incorrect return value")
}

// Ensures that isNil returns false for non-nil value.
func TestIsNilNotNilValue(t *testing.T) {
	assert := assert.New(t)
	assert.False(isNil(TestResource{}), "Incorrect return value")
}

// Ensures that isNil returns false for non-nil pointer.
func TestIsNilNotNilPtr(t *testing.T) {
	assert := assert.New(t)
	assert.False(isNil(&TestResource{}), "Incorrect return value")
}

// Ensures that NewRules panics if a non-pointer is passed in.
func TestNewRulesBadPtr(t *testing.T) {
	assert := assert.New(t)
	defer func() {
		r := recover()
		assert.NotNil(r, "Should have panicked")
	}()

	NewRules(struct{}{})
}

// Ensures that NewRules returns a Rules instance.
func TestNewRulesHappyPath(t *testing.T) {
	assert := assert.New(t)
	rule := &Rule{Field: "foo", FieldAlias: "bar"}

	rules := NewRules((*TestResource)(nil), rule)

	assert.Equal(rule, rules.Contents()[0])
	assert.Equal(reflect.TypeOf((*TestResource)(nil)).Elem(), rules.ResourceType())
}

// Ensures that Validate returns an error if a Rule doesn't have a Field or FieldAlias.
func TestRulesValidateNoName(t *testing.T) {
	assert := assert.New(t)
	rules := NewRules((*TestResource)(nil), &Rule{})

	assert.NotNil(rules.Validate())
}

// Ensures that Validate returns an error if a Rule type is not a struct or map.
func TestRulesValidateBadType(t *testing.T) {
	assert := assert.New(t)
	rules := NewRules((*string)(nil), &Rule{})

	assert.NotNil(rules.Validate())
}

// Ensures that Validate returns an error if a Rule's field type doesn't match the
// the resource field type.
func TestRulesValidateBadFieldType(t *testing.T) {
	assert := assert.New(t)
	rules := NewRules((*TestResource)(nil), &Rule{Field: "Foo", Type: Int})

	assert.NotNil(rules.Validate())
}

// Ensures that Validate returns an error if a Rule's field doesn't exist on the
// the resource.
func TestRulesValidateBadField(t *testing.T) {
	assert := assert.New(t)
	rules := NewRules((*TestResource)(nil), &Rule{Field: "Blah"})

	assert.NotNil(rules.Validate())
}

// Ensures that Validate does not return an error for non-resource Rules.
func TestRulesValidateNonResourceRule(t *testing.T) {
	assert := assert.New(t)
	rules := NewRules((*TestResource)(nil), &Rule{FieldAlias: "Blah"})

	assert.Nil(rules.Validate())
}
