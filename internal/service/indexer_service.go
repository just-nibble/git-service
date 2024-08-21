package service

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/just-nibble/git-service/internal/data"
	"github.com/just-nibble/git-service/pkg/github"
	"github.com/just-nibble/git-service/pkg/response"
)

type IndexerService struct {
	db           data.RepositoryStore
	githubClient *github.GitHubClient
}

func NewIndexerService(db data.RepositoryStore, gc *github.GitHubClient) *IndexerService {
	return &IndexerService{
		db:           db,
		githubClient: gc,
	}
}

func (s *IndexerService) AddRepository(w http.ResponseWriter, r *http.Request) {
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

	// Save repository in the database
	dbRepo := &data.Repository{
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

func (s *IndexerService) GetTopAuthors(w http.ResponseWriter, r *http.Request) {
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

	// Fetch top commit authors from the database
	authors, err := s.db.GetTopAuthors(n)
	if err != nil {
		http.Error(w, "Failed to retrieve authors", http.StatusInternalServerError)
		return
	}

	response.SuccessResponse(w, http.StatusOK, authors)
}

func (s *IndexerService) GetCommitsByRepo(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repo")
	if repoName == "" {
		http.Error(w, "Repository name is required", http.StatusBadRequest)
		return
	}

	// Fetch commits from the database
	commits, err := s.db.GetCommitsByRepository(repoName)
	if err != nil {
		http.Error(w, "Failed to retrieve commits", http.StatusInternalServerError)
		return
	}

	response.SuccessResponse(w, http.StatusOK, commits)
}

func (s *IndexerService) ResetStartDate(w http.ResponseWriter, r *http.Request) {
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
func (s *IndexerService) FetchAndSaveLatestCommits() {
	// Fetch all repositories
	repositories, err := s.db.GetAllRepositories()
	if err != nil {
		log.Printf("Failed to fetch repositories: %v", err)
		return
	}

	for _, repo := range repositories {
		// Fetch latest commits from GitHub
		commits, err := s.githubClient.GetCommits(repo.OwnerName, repo.Name, time.Now().Add(-time.Hour))
		if err != nil {
			log.Printf("Failed to fetch commits for repository %s: %v", repo.Name, err)
			continue
		}

		for _, commit := range commits {
			// Check if the commit already exists
			existingCommit, err := s.db.GetCommitByHash(commit.SHA)
			if err != nil && err.Error() != "commit not found" {
				log.Printf("Failed to check commit existence for hash %s: %v", commit.SHA, err)
				continue
			}

			if existingCommit != nil {
				// Commit already exists
				continue
			}

			// Retrieve or create the author
			author, err := s.db.GetOrCreateAuthor(commit.Commit.Author.Name, commit.Commit.Author.Email)
			if err != nil {
				log.Printf("Failed to retrieve or create author %s: %v", commit.Commit.Author.Name, err)
				continue
			}

			// Save the new commit
			newCommit := &data.Commit{
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
	}
}

// IndexCommits fetches and saves commits for a repository starting from the given date.
func (s *IndexerService) IndexCommits(repo *data.Repository) error {

	// Fetch latest commits from GitHub
	commits, err := s.githubClient.GetCommits(repo.OwnerName, repo.Name, repo.Since)
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

		if existingCommit != nil {
			// Commit already exists
			continue
		}

		// Retrieve or create the author
		author, err := s.db.GetOrCreateAuthor(commit.Commit.Author.Name, commit.Commit.Author.Email)
		if err != nil {
			log.Printf("Failed to retrieve or create author %s: %v", commit.Commit.Author.Name, err)
			continue
		}

		// Save the new commit
		newCommit := &data.Commit{
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

func (s *IndexerService) StartRepositoryMonitor(interval time.Duration) {
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

func (s *IndexerService) GetRepository(repoName string, repoOwner string) (github.Repository, error) {
	repo, err := s.githubClient.GetRepository(repoOwner, repoName)
	if err != nil {
		return github.Repository{}, err
	}

	return *repo, nil
}
