package rest

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensures that resourceHandlerProxy.Rules() injects the proxied ResourceHandler's
// resource type into the Rules.
func TestResourceHandlerProxyRules(t *testing.T) {
	assert := assert.New(t)
	handler := new(MockResourceHandler)
	proxy := resourceHandlerProxy{handler}
	handler.On("EmptyResource").Return(TestResource{})
	handler.On("Rules").Return(Rules{&Rule{}})

	rules := handler.Rules()
	assert.Equal(1, len(rules), "Rules should have 1 Rule")
	assert.Nil(rules[0].resourceType, "Rule's resourceType should be nil")

	rules = proxy.Rules()
	assert.Equal(1, len(rules), "Rules should have 1 Rule")
	assert.Equal(reflect.TypeOf(TestResource{}), rules[0].resourceType,
		"Incorrect resourceType")
}
