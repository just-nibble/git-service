package main

import (
	"log"
	"net/http"

	_ "github.com/just-nibble/git-service/cmd/indexer/docs" // Import your generated docs
	"github.com/just-nibble/git-service/internal/data"
	"github.com/just-nibble/git-service/internal/routes"
	"github.com/just-nibble/git-service/internal/service"
	"github.com/just-nibble/git-service/pkg/github"
)

// @title Github Service API
// @version 1.0
// @description This is an indexer for github.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host petstore.swagger.io
// @BasePath /
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

	// Start the HTTP server
	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}
