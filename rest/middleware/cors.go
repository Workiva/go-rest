package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/Workiva/go-rest/rest"
)

// NewCORSMiddleware returns a Middleware which enables cross-origin requests.
// Origin must match the supplied whitelist (which supports wildcards). Returns
// a MiddlewareError if the request should be terminated.
func NewCORSMiddleware(originWhitelist []string) rest.Middleware {
	return func(w http.ResponseWriter, r *http.Request) *rest.MiddlewareError {
		origin := r.Header.Get("Origin")
		if origin == "" && r.Method != "OPTIONS" {
			return nil
		}

		originMatch := false
		if checkOrigin(origin, originWhitelist) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header()["Access-Control-Allow-Headers"] = r.Header["Access-Control-Request-Headers"]
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			originMatch = true
		}

		if r.Method == "OPTIONS" {
			return &rest.MiddlewareError{Code: http.StatusOK}
		}

		if !originMatch {
			return &rest.MiddlewareError{
				Code:     http.StatusBadRequest,
				Response: []byte("Origin does not match whitelist"),
			}
		}

		return nil
	}
}

// checkOrigin checks if the given origin is contained in the origin whitelist.
// Returns true if the origin is in the whitelist, false if not.
func checkOrigin(origin string, whitelist []string) bool {
	url, err := url.Parse(origin)
	if err != nil {
		return false
	}
	originComponents := strings.Split(url.Host, ".")

checkWhitelist:
	for _, whitelisted := range whitelist {
		if whitelisted == "*" {
			return true
		}

		whitelistedComponents := strings.Split(whitelisted, ".")

		if len(originComponents) != len(whitelistedComponents) {
			// Do not match, try next host in whitelist.
			continue
		}

		for i, originComponent := range originComponents {
			whitelistedComponent := whitelistedComponents[i]
			if whitelistedComponent == "*" {
				// Wildcard, check next component.
				continue
			}

			if originComponent != whitelistedComponent {
				// Mismatch, try next host in whitelist.
				continue checkWhitelist
			}
		}

		// Origin matches whitelisted domain.
		return true
	}

	// No matches.
	return false
}
