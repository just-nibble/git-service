package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/just-nibble/git-service/internal/adapters/api"
	"github.com/just-nibble/git-service/internal/adapters/db"
	"github.com/just-nibble/git-service/internal/adapters/validators"
	"github.com/just-nibble/git-service/internal/core/domain/entities"
	"github.com/just-nibble/git-service/pkg/response"
)

type RepositoryService struct {
	rs           db.RepositoryStore
	cms          CommitService
	githubClient *api.GitHubClient
}

func NewRepositoryService(db db.RepositoryStore, cms CommitService, gc *api.GitHubClient) *RepositoryService {
	return &RepositoryService{
		rs:           db,
		githubClient: gc,
	}
}

func (s *RepositoryService) AddRepository(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Owner string `json:"owner"`
		Repo  string `json:"repo"`
		Since string `json:"since"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse the 'since' date
	sinceDate, err := time.Parse("2006-01-02", req.Since)
	if err != nil {
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid date format for 'since'")
		return
	}

	if req.Owner == "" || req.Repo == "" {
		http.Error(w, "Owner and repository name are required", http.StatusBadRequest)
		return
	}

	// Fetch repository details from GitHub
	repo, err := s.githubClient.GetRepository(req.Owner, req.Repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save repository in the dbbase
	dbRepo := &entities.Repository{
		OwnerName:       repo.Owner.Login,
		Name:            repo.Name,
		Description:     repo.Description,
		Language:        repo.Language,
		URL:             repo.URL,
		Since:           sinceDate,
		ForksCount:      repo.ForksCount,
		StarsCount:      repo.StarsCount,
		OpenIssuesCount: repo.OpenIssuesCount,
		WatchersCount:   repo.WatchersCount,
	}
	if err := s.rs.CreateRepository(dbRepo); err != nil {
		http.Error(w, "Failed to save repository", http.StatusInternalServerError)
		return
	}
	// Start background indexing of commits
	go func() {
		if err := s.cms.IndexCommits(dbRepo); err != nil {
			log.Println(err)
		}
	}()

	response.SuccessResponse(w, http.StatusCreated, dbRepo)
	log.Printf("Repository %s/%s added successfully", req.Owner, req.Repo)
}

func (s *RepositoryService) ResetStartDate(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repo")

	var req struct {
		Since string `json:"since"` // Expecting an ISO date string
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Parse the provided date
	since, err := time.Parse(time.RFC3339, req.Since)
	if err != nil {
		http.Error(w, "Invalid date format. Use RFC3339 format", http.StatusBadRequest)
		return
	}

	// Call the service to reset the start date
	if err := s.rs.ResetRepositoryStartDate(repoName, since); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Start date reset successfully"})
}

func (s *RepositoryService) StartRepositoryMonitor(interval time.Duration) {
	log.Println("monitoring...")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			repos, err := s.rs.GetAllRepositories()
			if err != nil {
				log.Println(err)
			}

			for _, repo := range repos {

				err = s.cms.IndexCommits(&repo)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func (s *RepositoryService) GetRepository(repoName string, repoOwner string) (api.Repository, error) {
	repo, err := s.githubClient.GetRepository(repoOwner, repoName)
	if err != nil {
		return api.Repository{}, err
	}

	return *repo, nil
}

// Seed seeds the database with the repository, along with its commits and authors, if the database is empty
func (s *RepositoryService) Seed(repo validators.Repo) error {
	if err := repo.Validate(); err != nil {
		return err
	}

	repoSlice := strings.Split(string(repo), "/")

	existingRepo, err := s.rs.GetRepositoryByName(repoSlice[1])
	if err != nil {
		return err
	}

	if existingRepo.ID == 0 {
		// Fetch repository details from GitHub
		repo, err := s.GetRepository(repoSlice[0], repoSlice[1])
		if err != nil {
			return err
		}

		// Save repository in the database
		seedRepo := &entities.Repository{
			OwnerName:       repo.Owner.Login,
			Name:            repo.Name,
			URL:             repo.URL,
			ForksCount:      repo.ForksCount,
			StarsCount:      repo.StarsCount,
			OpenIssuesCount: repo.OpenIssuesCount,
			WatchersCount:   repo.WatchersCount,
		}

		if err := s.rs.CreateRepository(seedRepo); err != nil {
			return err
		}

		// Start background indexing of commits
		go func() {
			if err := s.cms.IndexCommits(seedRepo); err != nil {
				log.Println(err)
			}
		}()

	}

	return nil
}

func (s *RepositoryService) ResumeIndexing() error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/rate_limit", nil)
	if err != nil {
		return fmt.Errorf("failed to create rate limit request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get rate limit: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch rate limit: status code %d", resp.StatusCode)
	}

	// Get the reset time from the headers
	resetTime := resp.Header.Get("X-RateLimit-Reset")
	resetTimestamp, err := strconv.ParseInt(resetTime, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse reset time: %v", err)
	}

	sleepDuration := time.Until(time.Unix(resetTimestamp, 0))
	log.Printf("Rate limit exceeded, sleeping for %v\n", sleepDuration)
	time.Sleep(sleepDuration)

	return nil // Ready to resume indexing
}
