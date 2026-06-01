package main

import (
	"log"
	"net/http"

	"github.com/yertaypert/go-assignment1/internal/handlers"
	"github.com/yertaypert/go-assignment1/internal/middleware"
)

func main() {
	log.SetFlags(0) // remove default log prefixes

	tasksHandler := http.HandlerFunc(handlers.TasksHandler)

	chained := middleware.Logging(middleware.APIKey(tasksHandler))

	http.Handle("/tasks", chained)
	http.ListenAndServe("127.0.0.1:8080", nil)
}
