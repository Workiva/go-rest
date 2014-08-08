package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensures that an empty Payload is returned by applyInboundRules if nil is passed in.
func TestApplyInboundRulesNilPayload(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(Payload{}, applyInboundRules(nil, []Rule{Rule{}}), "Incorrect return value")
}

// Ensures that payload is returned by applyInboundRules if rules is empty.
func TestApplyInboundRulesNoRules(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{}

	assert.Equal(payload, applyInboundRules(payload, []Rule{}), "Incorrect return value")
}

// Ensures that only inbound rules are applied and unspecified input fields are discarded.
func TestApplyInboundRules(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar", "baz": 1}
	rules := []Rule{
		Rule{
			Field:     "foo",
			ValueName: "foo",
		},
	}

	assert.Equal(
		Payload{"foo": "bar"},
		applyInboundRules(payload, rules),
		"Incorrect return value",
	)
}

// Ensures that inbound rules which specify an input handler yield the correct values.
func TestApplyInboundRulesInputHandler(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar", "baz": 1}
	rules := []Rule{
		Rule{
			Field:        "foo",
			ValueName:    "foo",
			InputHandler: func(val interface{}) interface{} { return val.(string) + " qux" },
		},
	}

	assert.Equal(
		Payload{"foo": "bar qux"},
		applyInboundRules(payload, rules),
		"Incorrect return value",
	)
}

// Ensures that nil is returned by applyOutboundRules if nil is passed in.
func TestApplyOutboundRulesNilResource(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(applyOutboundRules(nil, []Rule{Rule{}}), "Incorrect return value")
}

// Ensures that resource is returned by applyOutboundRules if rules is empty.
func TestApplyOutboundRulesNoRules(t *testing.T) {
	assert := assert.New(t)
	resource := &TestResource{}

	assert.Equal(resource, applyOutboundRules(resource, []Rule{}), "Incorrect return value")
}

// Ensures that resource is returned by applyOutboundRules if it's not a struct.
func TestApplyOutboundRulesNotStruct(t *testing.T) {
	assert := assert.New(t)
	resource := "resource"

	assert.Equal(resource, applyOutboundRules(resource, []Rule{Rule{}}), "Incorrect return value")
}

// Ensures that only outbound rules are applied and rules containing a field which doesn't exist
// are ignored.
func TestApplyOutboundRulesIgnoreBadRules(t *testing.T) {
	assert := assert.New(t)
	resource := &TestResource{Foo: "hello"}
	rules := []Rule{
		Rule{
			Field:     "Foo",
			ValueName: "f",
		},
		Rule{
			Field:     "Foo",
			ValueName: "FOO",
			InputOnly: true,
		},
		Rule{
			Field:     "bar",
			ValueName: "b",
		},
	}

	assert.Equal(
		Payload{"f": "hello"},
		applyOutboundRules(resource, rules),
		"Incorrect return value",
	)
}

// Ensures that rules which don't specify a valueName use the field name by default.
func TestApplyOutboundRulesDefaultName(t *testing.T) {
	assert := assert.New(t)
	resource := &TestResource{Foo: "hello"}
	rules := []Rule{
		Rule{
			Field: "Foo",
		},
	}

	assert.Equal(
		Payload{"Foo": "hello"},
		applyOutboundRules(resource, rules),
		"Incorrect return value",
	)
}

// Ensures that rules which specify an output Handler function yield the correct value.
func TestApplyOutboundRulesOutputHandler(t *testing.T) {
	assert := assert.New(t)
	resource := &TestResource{Foo: "hello"}
	rules := []Rule{
		Rule{
			Field:         "Foo",
			ValueName:     "foo",
			OutputHandler: func(val interface{}) interface{} { return val.(string) + " world" },
		},
	}

	assert.Equal(
		Payload{"foo": "hello world"},
		applyOutboundRules(resource, rules),
		"Incorrect return value",
	)
}
