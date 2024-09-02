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

type Indexer struct {
	db           db.RepositoryStore
	githubClient *api.GitHubClient
}

func NewIndexer(db db.RepositoryStore, gc *api.GitHubClient) *Indexer {
	return &Indexer{
		db:           db,
		githubClient: gc,
	}
}

func (s *Indexer) AddRepository(w http.ResponseWriter, r *http.Request) {
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
		URL:             repo.URL,
		Since:           sinceDate,
		ForksCount:      repo.ForksCount,
		StarsCount:      repo.StarsCount,
		OpenIssuesCount: repo.OpenIssuesCount,
		WatchersCount:   repo.WatchersCount,
	}
	if err := s.db.CreateRepository(dbRepo); err != nil {
		http.Error(w, "Failed to save repository", http.StatusInternalServerError)
		return
	}
	// Start background indexing of commits
	go func() {
		if err := s.IndexCommits(dbRepo); err != nil {
			log.Println(err)
		}
	}()

	response.SuccessResponse(w, http.StatusCreated, dbRepo)
	log.Printf("Repository %s/%s added successfully", req.Owner, req.Repo)
}

func (s *Indexer) GetTopAuthors(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repo")
	if repoName == "" {
		http.Error(w, "Repository name is required", http.StatusBadRequest)
		return
	}

	nStr := r.URL.Query().Get("n")
	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		http.Error(w, "Invalid number of authors", http.StatusBadRequest)
		return
	}

	// Fetch top commit authors from the dbbase
	authors, err := s.db.GetTopAuthors(n)
	if err != nil {
		http.Error(w, "Failed to retrieve authors", http.StatusInternalServerError)
		return
	}

	response.SuccessResponse(w, http.StatusOK, authors)
}

func (s *Indexer) GetCommitsByRepo(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repo")
	if repoName == "" {
		http.Error(w, "Repository name is required", http.StatusBadRequest)
		return
	}

	// Fetch commits from the dbbase
	commits, err := s.db.GetCommitsByRepository(repoName)
	if err != nil {
		http.Error(w, "Failed to retrieve commits", http.StatusInternalServerError)
		return
	}

	response.SuccessResponse(w, http.StatusOK, commits)
}

func (s *Indexer) ResetStartDate(w http.ResponseWriter, r *http.Request) {
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
	if err := s.db.ResetRepositoryStartDate(repoName, since); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Start date reset successfully"})
}

// FetchAndSaveLatestCommits fetches the latest commits from GitHub and saves them to the database
func (s *Indexer) FetchAndSaveLatestCommits() {
	page := 1
	perPage := 100

	// Fetch all repositories
	repositories, err := s.db.GetAllRepositories()
	if err != nil {
		log.Printf("Failed to fetch repositories: %v", err)
		return
	}

	for _, repo := range repositories {
		err := s.fetchAndSaveCommitsForRepo(repo, page, perPage)
		if err != nil {
			log.Printf("Failed to process repository %s: %v", repo.Name, err)
		}
	}
}

// fetchAndSaveCommitsForRepo handles fetching and saving commits for a single repository
func (s *Indexer) fetchAndSaveCommitsForRepo(repo entities.Repository, page, perPage int) error {
	for {
		// Fetch latest commits from GitHub with pagination and rate limiting checks
		commits, rateLimited, err := s.githubClient.GetCommits(
			repo.OwnerName, repo.Name, time.Now().Add(-time.Hour),
			page, perPage,
		)

		if err != nil {
			return fmt.Errorf("failed to fetch commits for repository %s: %v", repo.Name, err)
		}

		if rateLimited {
			s.ResumeIndexing()
		}

		for _, commit := range commits {
			err := s.processCommit(repo.ID, commit)
			if err != nil {
				log.Printf("Failed to process commit %s for repository %s: %v", commit.SHA, repo.Name, err)
			}
		}

		// Check if there are more pages of commits to fetch
		if len(commits) < perPage {
			break
		}

		page++
	}

	return nil
}

// processCommit handles the logic for checking the existence of a commit, retrieving/creating the author, and saving the commit
func (s *Indexer) processCommit(repoID uint, commit api.Commit) error {
	// Check if the commit already exists
	existingCommit, err := s.db.GetCommitByHash(commit.SHA)
	if err != nil && err.Error() != "commit not found" {
		return fmt.Errorf("failed to check commit existence for hash %s: %v", commit.SHA, err)
	}

	if existingCommit != nil {
		// Commit already exists
		return nil
	}

	// Retrieve or create the author
	author, err := s.db.GetOrCreateAuthor(commit.Commit.Author.Name, commit.Commit.Author.Email)
	if err != nil {
		return fmt.Errorf("failed to retrieve or create author %s: %v", commit.Commit.Author.Name, err)
	}

	// Save the new commit
	newCommit := &entities.Commit{
		RepositoryID: repoID,
		AuthorID:     author.ID,
		CommitHash:   commit.SHA,
		Message:      commit.Commit.Message,
		Date:         commit.Commit.Author.Date,
	}

	if err := s.db.CreateCommit(newCommit); err != nil {
		return fmt.Errorf("failed to save commit %s: %v", commit.SHA, err)
	}

	return nil
}

// IndexCommits fetches and saves commits for a repository starting from the given date.
func (s *Indexer) IndexCommits(repo *entities.Repository) error {
	page := 1
	perPage := 100
	// Fetch latest commits from GitHub
	commits, _, err := s.githubClient.GetCommits(
		repo.OwnerName, repo.Name, repo.Since,
		page, perPage,
	)

	if err != nil {
		log.Printf("Failed to fetch commits for repository %s: %v", repo.Name, err)
		return err
	}

	for _, commit := range commits {
		// Check if the commit already exists
		existingCommit, err := s.db.GetCommitByHash(commit.SHA)
		if err != nil && err.Error() != "commit not found" {
			log.Printf("Failed to check commit existence for hash %s: %v", commit.SHA, err)
			continue
		}

		if existingCommit.ID > 0 {
			// Commit already exists
			continue
		}

		// Retrieve or create the author
		author, err := s.db.GetOrCreateAuthor(commit.Commit.Author.Name, commit.Commit.Author.Email)
		if err != nil {
			continue
		}

		// Save the new commit
		newCommit := &entities.Commit{
			RepositoryID: repo.ID,
			AuthorID:     author.ID,
			CommitHash:   commit.SHA,
			Message:      commit.Commit.Message,
			Date:         commit.Commit.Author.Date,
		}

		if err := s.db.CreateCommit(newCommit); err != nil {
			log.Printf("Failed to save commit %s: %v", commit.SHA, err)
		}
	}

	return nil
}

func (s *Indexer) StartRepositoryMonitor(interval time.Duration) {
	log.Println("monitoring...")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			repos, err := s.db.GetAllRepositories()
			if err != nil {
				log.Println(err)
			}

			for _, repo := range repos {

				err = s.IndexCommits(&repo)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func (s *Indexer) GetRepository(repoName string, repoOwner string) (api.Repository, error) {
	repo, err := s.githubClient.GetRepository(repoOwner, repoName)
	if err != nil {
		return api.Repository{}, err
	}

	return *repo, nil
}

// Seed seeds the database with the repository, along with its commits and authors, if the database is empty
func (s *Indexer) Seed(repo validators.Repo) error {
	if err := repo.Validate(); err != nil {
		return err
	}

	repoSlice := strings.Split(string(repo), "/")

	existingRepo, err := s.db.GetRepositoryByName(repoSlice[1])
	if err != nil {
		return err
	}

	if existingRepo.ID == 0 {
		log.Println("Seeding database with Chromium repository...")

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

		if err := s.db.CreateRepository(seedRepo); err != nil {
			return err
		}

		// Start background indexing of commits
		go func() {
			if err := s.IndexCommits(seedRepo); err != nil {
				log.Println(err)
			}
		}()

	}

	return nil
}

func (s *Indexer) ResumeIndexing() error {
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
