package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensure that PathTrimMiddleware is correctly trimming the request path
func TestPathTrimMiddlewareTrimPrefix(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo/bar", nil)
	w := httptest.NewRecorder()

	assert.Nil(NewPathTrimMiddleware("/foo")(w, req))
	assert.Equal(req.URL.Path, "/bar")
}

// Ensure that PathTrimMiddleware correclty passes a path that doesn't need
// to be trimmed
func TestPathTrimMiddlewareNoTrimPrefix(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo/bar", nil)
	w := httptest.NewRecorder()

	assert.Nil(NewPathTrimMiddleware("/baz")(w, req))
	assert.Equal(req.URL.Path, "/foo/bar")
}
