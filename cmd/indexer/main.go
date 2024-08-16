package main

import (
	"log"
	"net/http"

	"github.com/just-nibble/git-service/internal/data"
	"github.com/just-nibble/git-service/internal/routes"
	"github.com/just-nibble/git-service/internal/service"
)

func main() {
	// Initialize the database
	db := data.InitDB()

	// Create the repository store
	repoStore := data.NewGormRepositoryStore(db)

	// Create the indexer service
	indexerService := service.NewIndexerService(repoStore)

	// Set up HTTP routes
	router := routes.NewRouter(indexerService)

	// Start the HTTP server
	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}
