package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timestamp := time.Now().Format("2006-01-02T15:04:05")
		method := r.Method
		path := r.URL.Path
		message := "request received"
		log.Printf("%s %s %s %s", timestamp, method, path, message)
		next.ServeHTTP(w, r)
	})
}
