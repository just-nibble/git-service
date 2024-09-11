package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/just-nibble/git-service/internal/adapters/api"
	"github.com/just-nibble/git-service/internal/adapters/http/dtos/routes"
	"github.com/just-nibble/git-service/internal/adapters/http/handlers"
	"github.com/just-nibble/git-service/internal/adapters/repository"
	"github.com/just-nibble/git-service/internal/adapters/storage"
	"github.com/just-nibble/git-service/internal/adapters/validators"
	"github.com/just-nibble/git-service/internal/core/service"
	"github.com/just-nibble/git-service/pkg/config"
)

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
	gc := api.NewGitHubClient(getenv("GITHUB_TOKEN", ""))

	// Create the repository store
	repoStore := repository.NewGormRepositoryStore(dB)

	authorStore := repository.NewGormAuthorStore(dB)
	authorService := service.NewAuthorService(authorStore, gc)

	commitStore := repository.NewGormCommitStore(dB)
	commitService := service.NewCommitService(
		commitStore, repoStore, authorStore, gc,
	)

	repoService := service.NewRepositoryService(
		repoStore, *commitService, gc,
	)

	repoHandler := handlers.NewRepositoryHandler(*repoService, *commitService)
	authorHandler := handlers.NewAuthorHandler(authorService)
	commitHandler := handlers.NewCommitHandler(*commitService)

	// Set up HTTP routes
	mux := http.NewServeMux()
	routes.NewAuthorRouter(mux, *authorHandler)
	routes.NewCommitRouter(mux, *commitHandler)
	routes.NewRepositoryRouter(mux, *repoHandler)

	interval, err := strconv.Atoi(getenv("MONITOR_INTERVAL", "60"))
	if err != nil {
		log.Fatal("Invalid MONITOR_INTERVAL VALUE")
	}

	if interval < 1 {
		log.Fatal("Please enter a number greater than zero as the MONITOR_INTERVAL value")
	}

	start_date, err := time.Parse(time.RFC3339, getenv("DEFAULT_START_DATE", "2012-03-06T23:06:50Z"))
	if err != nil {
		log.Println(err)
		log.Fatal("Invalid default start date")
	}

	cfg := config.Config{
		DefaultRepository: validators.Repo(getenv("DEFAULT_REPO", "chromium/chromium")),
		DefaultStartDate:  start_date,
		MonitorInterval:   interval,
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Seed the database if necessary
	if err := repoService.Seed(ctx, cfg); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Start the background worker
	go repoService.StartRepositoryMonitor(ctx, time.Duration(cfg.MonitorInterval)*time.Hour)

	// Start the HTTP server
	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
