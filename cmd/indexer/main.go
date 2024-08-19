package main

import (
	"log"
	"net/http"
	"time"

	"github.com/just-nibble/git-service/internal/data"
	"github.com/just-nibble/git-service/internal/routes"
	"github.com/just-nibble/git-service/internal/seeder"
	"github.com/just-nibble/git-service/internal/service"
	"github.com/just-nibble/git-service/pkg/github"
)

func main() {
	// Initialize the database
	db := data.InitDB()

	// Create the repository store
	repoStore := data.NewGormRepositoryStore(db)

	// Initialize the GitHub client
	gc := github.NewGitHubClient()

	// Create the indexer service
	indexerService := service.NewIndexerService(repoStore, gc)

	// Set up HTTP routes
	router := routes.NewRouter(indexerService)

	// Seed the database if necessary
	if err := seeder.SeedDatabase(db, indexerService); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Start the background worker
	go indexerService.StartRepositoryMonitor(1 * time.Minute)

	// Start the HTTP server
	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}
