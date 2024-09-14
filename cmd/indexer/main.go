package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/just-nibble/git-service/internal/http/dtos"
	"github.com/just-nibble/git-service/internal/http/handlers"
	"github.com/just-nibble/git-service/internal/http/routes"
	"github.com/just-nibble/git-service/internal/repository"
	"github.com/just-nibble/git-service/internal/usecases"
	"github.com/just-nibble/git-service/pkg/config"
	"github.com/just-nibble/git-service/pkg/database"
	"github.com/just-nibble/git-service/pkg/errcodes"
	"github.com/just-nibble/git-service/pkg/git"
	"github.com/just-nibble/git-service/pkg/log"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	log := log.NewLogger()
	config, err := config.LoadConfig(*log)
	if err != nil {
		log.Error.Printf("failed to load config %s", err.Error())
	}

	// Create a new PostgresDatabase instance
	dbClient := database.NewPostgresDatabase(config.DSN, 10, 5, 3*time.Hour)
	err = dbClient.ConnectDB(ctx)
	if err != nil {
		log.Error.Printf("failed to establish postgres database connection: %s", err.Error())
	}

	// Run database migrations
	if err := dbClient.Migrate(ctx); err != nil {
		log.Error.Printf("failed to run database migrations: %s", err.Error())
	}

	githubClient := git.NewGitHubClient(config.GitClientBaseURL, config.GitClientToken, config.MonitorInterval)

	dB := dbClient.GetDB()

	repoRepository := repository.NewGormRepositoryMetaRepository(dB)
	authorRepository := repository.NewGormAuthorRepository(dB)
	commitRepository := repository.NewGormCommitRepository(dB)

	commitUsecase := usecases.NewGitCommitUsecase(commitRepository, repoRepository)
	gitRepoUsecase := usecases.NewrepoMetaUsecase(repoRepository, commitRepository, authorRepository, githubClient, *config, *log)
	authorUsecase := usecases.NewAuthorUseCase(authorRepository)

	repoHandler := handlers.NewRepositoryHandler(gitRepoUsecase)
	authorHandler := handlers.NewAuthorHandler(authorUsecase)
	commitHandler := handlers.NewCommitHandler(commitUsecase)

	// Set up HTTP routes
	mux := http.NewServeMux()
	routes.NewAuthorRouter(mux, *authorHandler)
	routes.NewCommitRouter(mux, *commitHandler)
	routes.NewRepositoryRouter(mux, *repoHandler)

	err = seedDefaultRepository(config, gitRepoUsecase, *log)
	if err != nil && err != errcodes.ErrRepoAlreadyAdded {
		log.Error.Fatalf("failed to seed default repository: %s,", err.Error())
	}

	go gitRepoUsecase.ResumeIndexing(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info.Println("Program is shutting down...")
				// Call method to set isFetching to false in DB
				if err := gitRepoUsecase.ModifyRepoStatus(context.Background(), false); err != nil {
					log.Error.Printf("Error updating index to false: %s", err.Error())
				}
				os.Exit(0)
			default:
				time.Sleep(5 * time.Second)
			}
		}
	}()

	// Start the HTTP server
	log.Info.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Error.Fatalf("Could not start server: %v", err)
	}
}

// seedDefaultRepository seeds a default repository to database
func seedDefaultRepository(config *config.Config, repositoryUsecase usecases.RepoMetaUsecase, log log.Log) error {
	defaultRepo := dtos.RepositoryInput{
		Name: config.DefaultRepository,
	}

	repo, err := repositoryUsecase.InitiateIndexing(context.Background(), defaultRepo)
	if err != nil && err != errcodes.ErrNoRecordFound {
		return err
	}

	if repo != nil {
		log.Info.Printf("Successfully seeded default repository: %s", repo.Name)
	}
	return err
}
