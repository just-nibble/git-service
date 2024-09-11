package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/just-nibble/git-service/internal/adapters/api"
	"github.com/just-nibble/git-service/internal/adapters/repository"
	"github.com/just-nibble/git-service/internal/core/domain/entities"
	"github.com/just-nibble/git-service/pkg/config"
)

type RepositoryService struct {
	rs           repository.RepositoryStore
	cms          CommitService
	githubClient *api.GitHubClient
}

func NewRepositoryService(rs repository.RepositoryStore, cms CommitService, gc *api.GitHubClient) *RepositoryService {
	return &RepositoryService{
		rs:           rs,
		cms:          cms,
		githubClient: gc,
	}
}

func (s *RepositoryService) StartRepositoryMonitor(ctx context.Context, interval time.Duration) {
	log.Println("monitoring...")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cms.FetchAndSaveLatestCommits(ctx)
		case <-ctx.Done():
			log.Println("Repository monitor stopped.")
			return // Stop the monitor when context is canceled
		}

	}
}

func (s *RepositoryService) CreateRepository(repoOwner string, repoName string, since time.Time) (*entities.RepositoryMeta, error) {
	// Fetch repository details from GitHub
	repo, err := s.githubClient.GetRepository(repoOwner, repoName, since)
	if err != nil {
		return &entities.RepositoryMeta{}, err
	}

	// Start background indexing of commits

	// Save repository in the dbbase
	dbRepo := &repository.Repository{
		OwnerName:       repo.OwnerName,
		Name:            repo.Name,
		Description:     repo.Description,
		Language:        repo.Language,
		URL:             repo.URL,
		Since:           repo.Since,
		ForksCount:      repo.ForksCount,
		StarsCount:      repo.StarsCount,
		OpenIssuesCount: repo.OpenIssuesCount,
		WatchersCount:   repo.WatchersCount,
		Index:           true,
	}

	if err := s.rs.SaveRepository(dbRepo); err != nil {
		return nil, err
	}

	return repo, nil
}

func (s *RepositoryService) GetRepository(repoOwner string, repoName string, since time.Time) (entities.RepositoryMeta, error) {
	repo, err := s.githubClient.GetRepository(repoOwner, repoName, since)
	if err != nil {
		return entities.RepositoryMeta{}, err
	}

	return *repo, nil
}

func (s *RepositoryService) ResetRepositoryStartDate(repoName string, since time.Time) error {
	if err := s.rs.ResetRepositoryStartDate(repoName, since); err != nil {
		return err
	}

	return nil
}

// Seed seeds the database with the repository, along with its commits and authors, if the database is empty
func (s *RepositoryService) Seed(ctx context.Context, cfg config.Config) error {
	if err := cfg.DefaultRepository.Validate(); err != nil {
		return err
	}

	repoSlice := strings.Split(string(cfg.DefaultRepository), "/")

	existingRepo, err := s.rs.GetRepositoryByName(repoSlice[1])
	if err != nil {
		return err
	}

	if existingRepo.ID == 0 {
		// Fetch repository details from GitHub
		repo, err := s.GetRepository(repoSlice[0], repoSlice[1], cfg.DefaultStartDate)
		if err != nil {
			return err
		}

		// Save repository in the database
		seedRepo := &repository.Repository{
			OwnerName:       repo.OwnerName,
			Name:            repo.Name,
			Description:     repo.Description,
			Language:        repo.Language,
			URL:             repo.URL,
			CreatedAt:       repo.CreatedAt,
			UpdatedAt:       repo.UpdatedAt,
			ForksCount:      repo.ForksCount,
			StarsCount:      repo.StarsCount,
			OpenIssuesCount: repo.OpenIssuesCount,
			WatchersCount:   repo.WatchersCount,
			Index:           true,
			LastPage:        1,
		}

		if err := s.rs.SaveRepository(seedRepo); err != nil {
			return err
		}

		repo.ID = seedRepo.ID

		// Start background indexing of commits
		go func() {
			log.Println("Seeding...")
			s.cms.IndexCommits(ctx, &repo)
		}()

	}

	return nil
}

func (s *RepositoryService) ResumeIndexing() error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/rate_limit", nil)
	if err != nil {
		return fmt.Errorf("failed to create rate limit request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get rate limit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch rate limit: status code %d", resp.StatusCode)
	}

	// Get the reset time from the headers
	resetTime := resp.Header.Get("X-RateLimit-Reset")
	resetTimestamp, err := strconv.ParseInt(resetTime, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse reset time: %w", err)
	}

	sleepDuration := time.Until(time.Unix(resetTimestamp, 0))
	log.Printf("Rate limit exceeded, sleeping for %v\n", sleepDuration)
	time.Sleep(sleepDuration)

	return nil // Ready to resume indexing
}

func (s *RepositoryService) UpdateRepository(repo *entities.RepositoryMeta) (*entities.RepositoryMeta, error) {

	// Save repository in the dbbase
	dbRepo := &repository.Repository{
		OwnerName:       repo.OwnerName,
		Name:            repo.Name,
		Description:     repo.Description,
		Language:        repo.Language,
		URL:             repo.URL,
		Since:           repo.Since,
		ForksCount:      repo.ForksCount,
		StarsCount:      repo.StarsCount,
		OpenIssuesCount: repo.OpenIssuesCount,
		WatchersCount:   repo.WatchersCount,
		Index:           repo.Index,
	}

	if err := s.rs.SaveRepository(dbRepo); err != nil {
		return nil, err
	}

	return repo, nil
}

func (s *RepositoryService) GetRepositoryByName(name string) (*entities.RepositoryMeta, error) {
	repo, err := s.rs.GetRepositoryByName(name)
	if err != nil {
		return nil, err
	}

	return repo, nil
}
