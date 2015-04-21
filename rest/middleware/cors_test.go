package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensures that CORSMiddleware applies the headers needed for CORS and returns
// true for non-OPTIONS requests.
func TestCORSMiddleware(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Set("Origin", "abc")
	req.Header.Set("Access-Control-Request-Headers", "def")
	w := httptest.NewRecorder()
	assert.False(CORSMiddleware(w, req))

	assert.Equal("abc", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal("POST, GET, OPTIONS, PUT, DELETE", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal([]string{"def"}, w.Header()["Access-Control-Allow-Headers"])
	assert.Equal([]string{"true"}, w.Header()["Access-Control-Allow-Credentials"])
}

// Ensures that CORSMiddleware returns true for OPTIONS requests.
func TestCORSMiddlewareOptionsRequest(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "http://example.com/foo", nil)
	req.Header.Set("Origin", "abc")
	w := httptest.NewRecorder()
	assert.True(t, CORSMiddleware(w, req))
}
