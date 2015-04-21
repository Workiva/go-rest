package middleware

import "net/http"

// CORSMiddleware enables cross-origin requests. It implements the Middleware
// interface.
func CORSMiddleware(w http.ResponseWriter, r *http.Request) bool {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header()["Access-Control-Allow-Headers"] = r.Header["Access-Control-Request-Headers"]
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	return r.Method == "OPTIONS"
}
