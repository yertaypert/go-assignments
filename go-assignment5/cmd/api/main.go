package main

import (
	"log"
	"net/http"

	"github.com/yertaypert/go-assignment5/database"
	"github.com/yertaypert/go-assignment5/handlers"
	"github.com/yertaypert/go-assignment5/repository"
)

func main() {
	db := database.Connect()
	repo := repository.NewRepository(db)
	handler := handlers.NewHandler(repo)

	http.HandleFunc("/users", handler.GetPaginatedUsers)
	http.HandleFunc("/users/common-friends", handler.GetCommonFriends)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
