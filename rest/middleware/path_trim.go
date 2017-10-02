package middleware

import(
	"net/http"
	"strings"

	"github.com/Workiva/go-rest/rest"
)

// NewPathTrimMiddleware returns a Middleware which inspects request paths
// and removes the content of trimPortion from the prefix. With a path of /foo/bar
// and a trimpPortion of /foo you would be left with a path of /bar.
func NewPathTrimMiddleware(trimPortion string) rest.Middleware {
	return func(w http.ResponseWriter, r *http.Request) *rest.MiddlewareError {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, trimPortion)
		return nil
	}
}
