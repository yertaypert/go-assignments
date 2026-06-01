package main

import (
	"log"

	"github.com/yertaypert/go-assignment7/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
