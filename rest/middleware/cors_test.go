package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensures that CORSMiddleware applies the headers needed for CORS when * is
// present in the whitelist.
func TestCORSMiddlewareAll(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Set("Origin", "http://foo.com")
	req.Header.Set("Access-Control-Request-Headers", "def")
	w := httptest.NewRecorder()
	assert.Nil(NewCORSMiddleware([]string{"*"})(w, req))

	assert.Equal("http://foo.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal("POST, GET, OPTIONS, PUT, DELETE", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal([]string{"def"}, w.Header()["Access-Control-Allow-Headers"])
	assert.Equal([]string{"true"}, w.Header()["Access-Control-Allow-Credentials"])
}

// Ensures that CORSMiddleware applies the headers needed for CORS and respects
// the origin whitelist.
func TestCORSMiddlewareWhitelist(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Set("Origin", "http://foo.wdesk.com")
	req.Header.Set("Access-Control-Request-Headers", "def")
	w := httptest.NewRecorder()
	middleware := NewCORSMiddleware([]string{"blah.wdesk.org", "*.wdesk.com"})
	assert.Nil(middleware(w, req))

	assert.Equal("http://foo.wdesk.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal("POST, GET, OPTIONS, PUT, DELETE", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal([]string{"def"}, w.Header()["Access-Control-Allow-Headers"])
	assert.Equal([]string{"true"}, w.Header()["Access-Control-Allow-Credentials"])

	// Mismatched origin
	req, _ = http.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Set("Origin", "http://baz.wdesk.org")
	req.Header.Set("Access-Control-Request-Headers", "def")
	w = httptest.NewRecorder()
	err := middleware(w, req)
	assert.Error(err)
	assert.Equal(http.StatusBadRequest, err.Code())

	assert.Equal("", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal("", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Nil(w.Header()["Access-Control-Allow-Headers"])
	assert.Nil(w.Header()["Access-Control-Allow-Credentials"])
}

// Ensures that CORSMiddleware returns a MiddlewareError with a 200 response
// code.
func TestCORSMiddlewareOptionsRequest(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "http://example.com/foo", nil)
	req.Header.Set("Origin", "http://foo.com")
	w := httptest.NewRecorder()
	err := NewCORSMiddleware([]string{"foo.com"})(w, req)
	assert.Equal(t, http.StatusOK, err.Code())
}
