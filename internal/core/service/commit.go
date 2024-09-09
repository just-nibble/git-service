package service

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/just-nibble/git-service/internal/adapters/api"
	"github.com/just-nibble/git-service/internal/adapters/db"
	"github.com/just-nibble/git-service/internal/core/domain/entities"
	"github.com/just-nibble/git-service/pkg/response"
)

type CommitService struct {
	cs           db.CommitStore
	rs           db.RepositoryStore
	as           db.AuthorStore
	githubClient *api.GitHubClient
}

func NewCommitService(db db.CommitStore, rs db.RepositoryStore, as db.AuthorStore, gc *api.GitHubClient) *CommitService {
	return &CommitService{
		cs:           db,
		githubClient: gc,
	}
}

func (s *CommitService) GetCommitsByRepo(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repo")
	if repoName == "" {
		http.Error(w, "Repository name is required", http.StatusBadRequest)
		return
	}

	// Fetch commits from the dbbase
	commits, err := s.cs.GetCommitsByRepository(repoName)
	if err != nil {
		http.Error(w, "Failed to retrieve commits", http.StatusInternalServerError)
		return
	}

	response.SuccessResponse(w, http.StatusOK, commits)
}

// FetchAndSaveLatestCommits fetches the latest commits from GitHub and saves them to the database
func (s *CommitService) FetchAndSaveLatestCommits() {
	page := 1
	perPage := 100

	// Fetch all repositories
	repositories, err := s.rs.GetAllRepositories()
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
func (s *CommitService) fetchAndSaveCommitsForRepo(repo entities.Repository, page, perPage int) error {
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
func (s *CommitService) processCommit(repoID uint, commit api.Commit) error {
	// Check if the commit already exists
	existingCommit, err := s.cs.GetCommitByHash(commit.SHA)
	if err != nil && err.Error() != "commit not found" {
		return fmt.Errorf("failed to check commit existence for hash %s: %v", commit.SHA, err)
	}

	if existingCommit != nil {
		// Commit already exists
		return nil
	}

	// Retrieve or create the author
	author, err := s.as.GetOrCreateAuthor(commit.Commit.Author.Name, commit.Commit.Author.Email)
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

	if err := s.cs.CreateCommit(newCommit); err != nil {
		return fmt.Errorf("failed to save commit %s: %v", commit.SHA, err)
	}

	return nil
}

// IndexCommits fetches and saves commits for a repository starting from the given date.
func (s *CommitService) IndexCommits(repo *entities.Repository) error {
	page := 1
	perPage := 100
	// Fetch latest commits from GitHub
	commits, _, err := s.githubClient.GetCommits(
		repo.OwnerName, repo.Name, repo.Since,
		page, perPage,
	)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		// Check if the commit already exists
		existingCommit, err := s.cs.GetCommitByHash(commit.SHA)
		if err != nil && err.Error() != "commit not found" {
			continue
		}

		if existingCommit.ID > 0 {
			// Commit already exists
			continue
		}

		// Retrieve or create the author
		author, err := s.as.GetOrCreateAuthor(commit.Commit.Author.Name, commit.Commit.Author.Email)
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

		if err := s.cs.CreateCommit(newCommit); err != nil {
			log.Printf("Failed to save commit %s: %v", commit.SHA, err)
		}
	}

	return nil
}

func (s *CommitService) ResumeIndexing() error {
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
