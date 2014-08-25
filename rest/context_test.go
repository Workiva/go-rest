package rest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensures that if a limit doesn't exist on the context, the default is returned.
func TestLimitDefault(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	ctx := NewContext(nil, req)
	assert.Equal(100, ctx.Limit())
}

// Ensures that if an invalid limit value is on the context, the default is returned.
func TestLimitBadValue(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	ctx := NewContext(nil, req)
	ctx = ctx.WithValue(limitKey, "blah")
	assert.Equal(100, ctx.Limit())
}

// Ensures that the correct limit is returned from the context.
func TestLimit(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	ctx := NewContext(nil, req)
	ctx = ctx.WithValue(limitKey, "5")
	assert.Equal(5, ctx.Limit())
}
