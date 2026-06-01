package middleware

import "net/http"

// APIKey returns middleware that checks for valid X-API-KEY header
func APIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")
		if apiKey != "secret12345" {
			sendError(w, http.StatusUnauthorized, "unauthorized, invalid API key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error": "` + message + `"}`))
}
