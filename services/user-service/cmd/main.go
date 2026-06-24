package main

import (
	"log"
	"net/http"

	"github.com/sumithsudesan/user-service/src/user"
)

func main() {
	// Create a new instance of the user service.
	svc := user.NewService()

	// Create a new handler for the user service.
	handler := user.NewHandler(svc)

	// Create a new router for the user service.
	router := user.NewRouter(handler)

	// Start the HTTP server on port 8080 and log any fatal errors.
	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
