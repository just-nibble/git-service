package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/just-nibble/git-service/internal/adapters/api"
	"github.com/just-nibble/git-service/internal/adapters/db"
	routes "github.com/just-nibble/git-service/internal/adapters/http"
	"github.com/just-nibble/git-service/internal/adapters/storage"
	"github.com/just-nibble/git-service/internal/adapters/validators"
	"github.com/just-nibble/git-service/internal/core/service"
)

type Config struct {
	defaultRepository validators.Repo `validate:"required"`
	monitorInterval   int
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func main() {
	// Initialize the database
	dB := storage.InitDB()
	// Initialize the GitHub client
	gc := api.NewGitHubClient()

	// Create the repository store
	repoStore := db.NewGormRepositoryStore(dB)

	authorStore := db.NewGormAuthorStore(dB)
	authorService := service.NewAuthorService(authorStore, gc)

	commitStore := db.NewGormCommitStore(dB)
	commitService := service.NewCommitService(
		commitStore, repoStore, authorStore, gc,
	)

	repoService := service.NewRepositoryService(
		repoStore, *commitService, gc,
	)

	// Set up HTTP routes
	mux := http.NewServeMux()
	routes.NewAuthorRouter(mux, authorService)
	routes.NewCommitRouter(mux, commitService)
	routes.NewRepositoryRouter(mux, repoService)

	interval, err := strconv.Atoi(getenv("MONITOR_INTERVAL", "60"))
	if err != nil {
		log.Fatal("Invalid MONITOR_INTERVAL VALUE")
	}

	if interval < 1 {
		log.Fatal("Please enter a number greater than zero as the MONITOR_INTERVAL value")
	}

	cfg := Config{
		defaultRepository: validators.Repo(getenv("DEFAULT_REPO", "chromium/chromium")),
		monitorInterval:   interval,
	}

	// Seed the database if necessary
	if err := repoService.Seed(cfg.defaultRepository); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Start the background worker
	go repoService.StartRepositoryMonitor(time.Duration(cfg.monitorInterval) * time.Minute)

	// Start the HTTP server
	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
