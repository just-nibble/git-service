package service

import (
	"context"
	"errors"
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
	cs db.CommitStore
	rs db.RepositoryStore
	as db.AuthorStore
	gc *api.GitHubClient
}

func NewCommitService(cs db.CommitStore, rs db.RepositoryStore, as db.AuthorStore, gc *api.GitHubClient) *CommitService {
	return &CommitService{
		cs: cs,
		rs: rs,
		as: as,
		gc: gc,
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
func (s *CommitService) FetchAndSaveLatestCommits(ctx context.Context) {
	page := 1
	perPage := 100

	// Fetch all repositories
	repositories, err := s.rs.GetAllRepositories()
	if err != nil {
		log.Printf("Failed to fetch repositories: %v", err)
		return
	}

	for _, repo := range repositories {
		err := s.fetchAndSaveCommitsForRepo(ctx, repo, page, perPage)
		if err != nil {
			log.Printf("Failed to process repository %s: %v", repo.Name, err)
		}
	}
}

// fetchAndSaveCommitsForRepo handles fetching and saving commits for a single repository
func (s *CommitService) fetchAndSaveCommitsForRepo(ctx context.Context, repo entities.Repository, page, perPage int) error {
	for {
		// Fetch latest commits from GitHub with pagination and rate limiting checks
		commits, err := s.gc.GetCommits(
			repo.OwnerName, repo.Name, time.Now().Add(-time.Hour),
			page, perPage,
		)
		if err != nil {
			if err.Error() == "rate limited" {
				s.ResumeIndexing()
			}
			return fmt.Errorf("failed to fetch commits for repository %s: %w", repo.Name, err)
		}

		for _, commit := range commits {
			err := s.processCommit(ctx, repo.ID, commit)
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
func (s *CommitService) processCommit(ctx context.Context, repoID uint, commit api.Commit) error {
	// Check if the commit already exists
	existingCommit, err := s.cs.GetCommitByHash(commit.SHA)
	if err != nil && err.Error() != "commit not found" {
		return fmt.Errorf("failed to check commit existence for hash %s: %w", commit.SHA, err)
	}

	if existingCommit != nil {
		// Commit already exists
		return nil
	}
	// Retrieve or create the author
	author, err := s.as.GetOrCreateAuthor(ctx, commit.Commit.Author.Name, commit.Commit.Author.Email)
	if err != nil {
		return fmt.Errorf("failed to retrieve or create author %s: %w", commit.Commit.Author.Name, err)
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
		return fmt.Errorf("failed to save commit %s: %w", commit.SHA, err)
	}

	return nil
}

// IndexCommits fetches and saves commits for a repository starting from the given date.
func (s *CommitService) IndexCommits(ctx context.Context, repo *entities.Repository) error {
	if s.gc == nil {
		return errors.New("GitHub client is not initialized")
	}
	page := 1
	perPage := 100

	for {
		commits, err := s.gc.GetCommits(
			repo.OwnerName, repo.Name, repo.Since,
			page, perPage,
		)

		if err != nil {
			if err.Error() == "rate limited" {
				if err := s.ResumeIndexing(); err != nil {
					return err
				}
				// After resuming, refetch commits for the same page
				continue
			}
			return err
		}

		// Process fetched commits
		for _, commit := range commits {
			// Check if the commit already exists
			existingCommit, err := s.cs.GetCommitByHash(commit.SHA)
			if err != nil && err.Error() != "commit not found" {
				continue
			}

			if existingCommit.ID > 0 {
				continue // Commit already exists
			}

			// Retrieve or create the author
			author, err := s.as.GetOrCreateAuthor(ctx, commit.Commit.Author.Name, commit.Commit.Author.Email)
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

			log.Printf("Saved commit #%s for %s", commit.SHA, repo.Name)
		}

		// Break if the commit list is smaller than `perPage`, meaning no more commits
		if len(commits) < perPage {
			break
		}
		page++
	}

	return nil
}

func (s *CommitService) ResumeIndexing() error {
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
