package middleware

import (
	"log"
	"net/http"
	"os"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		validAPIKey := os.Getenv("API_KEY")
		apiKey := r.Header.Get("X-API-KEY")

		if apiKey == "" {
			log.Println("missing API key")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if apiKey != validAPIKey {
			log.Println("invalid API key")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
