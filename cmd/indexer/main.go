package main

import (
	"log"
	"net/http"
	"time"

	"github.com/just-nibble/git-service/internal/adapters/api"
	"github.com/just-nibble/git-service/internal/adapters/db"
	routes "github.com/just-nibble/git-service/internal/adapters/http"
	"github.com/just-nibble/git-service/internal/adapters/storage"
	"github.com/just-nibble/git-service/internal/core/service"
)

func main() {
	// Initialize the database
	dB := storage.InitDB()

	// Create the repository store
	repoStore := db.NewGormRepositoryStore(dB)

	// Initialize the GitHub client
	gc := api.NewGitHubClient()

	// Create the indexer service
	indexer := service.NewIndexer(repoStore, gc)

	// Set up HTTP routes
	router := routes.NewRouter(indexer)
	// Seed the database if necessary
	if err := indexer.Seed(); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Start the background worker
	go indexer.StartRepositoryMonitor(1 * time.Minute)

	// Start the HTTP server
	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}
