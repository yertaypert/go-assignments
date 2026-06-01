package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf(
			"timestamp=%s method=%s endpoint=%s duration=%s",
			start.Format(time.RFC3339),
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}
