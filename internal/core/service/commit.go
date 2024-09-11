package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/just-nibble/git-service/internal/adapters/api"
	"github.com/just-nibble/git-service/internal/adapters/repository"
	"github.com/just-nibble/git-service/internal/core/domain/entities"
)

type CommitService struct {
	cs repository.CommitStore
	rs repository.RepositoryStore
	as repository.AuthorStore
	gc *api.GitHubClient
}

func NewCommitService(cs repository.CommitStore, rs repository.RepositoryStore, as repository.AuthorStore, gc *api.GitHubClient) *CommitService {
	return &CommitService{
		cs: cs,
		rs: rs,
		as: as,
		gc: gc,
	}
}

func (s *CommitService) GetCommitsByRepo(repo string) ([]entities.Commit, error) {
	// Fetch commits from the dbbase
	cs, err := s.cs.GetCommitsByRepository(repo)
	if err != nil {
		return nil, err
	}

	var commits []entities.Commit

	for _, v := range cs {
		commit := entities.Commit{
			ID:         v.ID,
			CommitHash: v.CommitHash,
			Message:    v.Message,
			Date:       v.Date,
			Author: entities.Author{
				ID:    v.Author.ID,
				Name:  v.Author.Name,
				Email: v.Author.Email,
			},
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

// FetchAndSaveLatestCommits fetches the latest commits from GitHub and saves them to the database
func (s *CommitService) FetchAndSaveLatestCommits(ctx context.Context) {
	perPage := 100

	// Fetch all repositories
	repositories, err := s.rs.GetAllRepositories()
	if err != nil {
		log.Printf("Failed to fetch repositories: %v", err)
		return
	}

	for _, repo := range repositories {
		if !repo.Index {
			page := repo.LastPage
			err := s.fetchAndSaveCommitsForRepo(ctx, repo, page, perPage)
			if err != nil {
				log.Printf("Failed to process repository %s: %v", repo.Name, err)
			}
		}
	}
}

// fetchAndSaveCommitsForRepo handles fetching and saving commits for a single repository
func (s *CommitService) fetchAndSaveCommitsForRepo(ctx context.Context, repo entities.RepositoryMeta, page, perPage int) error {
	for {
		// Fetch latest commits from GitHub with pagination and rate limiting checks
		commits, err := s.gc.GetCommits(
			repo.OwnerName, repo.Name, repo.Since,
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
				log.Printf("Failed to process commit %s for repository %s: %v", commit.CommitHash, repo.Name, err)
			}
		}

		// Check if there are more pages of commits to fetch
		if len(commits) < perPage {
			break
		}

		page++
	}

	dbRepo := &repository.Repository{
		ID:              repo.ID,
		OwnerName:       repo.OwnerName,
		Name:            repo.Name,
		Description:     repo.Description,
		Language:        repo.Language,
		URL:             repo.URL,
		ForksCount:      repo.ForksCount,
		StarsCount:      repo.StarsCount,
		OpenIssuesCount: repo.OpenIssuesCount,
		WatchersCount:   repo.WatchersCount,
		CreatedAt:       repo.CreatedAt,
		UpdatedAt:       repo.UpdatedAt,
		LastPage:        page,
	}
	s.rs.SaveRepository(dbRepo)

	return nil
}

// processCommit handles the logic for checking the existence of a commit, retrieving/creating the author, and saving the commit
func (s *CommitService) processCommit(ctx context.Context, repoID uint, commit entities.Commit) error {
	// Check if the commit already exists
	existingCommit, err := s.cs.GetCommitByHash(commit.CommitHash)
	if err != nil && err.Error() != "commit not found" {
		return fmt.Errorf("failed to check commit existence for hash %s: %w", commit.CommitHash, err)
	}

	if existingCommit != nil {
		// Commit already exists
		return nil
	}
	// Retrieve or create the author
	author, err := s.as.GetOrCreateAuthor(ctx, commit.Author.Name, commit.Author.Email)
	if err != nil {
		return fmt.Errorf("failed to retrieve or create author %s: %w", commit.Author.Name, err)
	}

	// Save the new commit
	newCommit := &repository.Commit{
		RepositoryID: repoID,
		AuthorID:     author.ID,
		CommitHash:   commit.CommitHash,
		Message:      commit.Message,
		Date:         commit.Date,
	}

	if err := s.cs.CreateCommit(newCommit); err != nil {
		return fmt.Errorf("failed to save commit %s: %w", commit.CommitHash, err)
	}

	return nil
}

// IndexCommits fetches and saves commits for a repository starting from the given date.
func (s *CommitService) IndexCommits(ctx context.Context, repo *entities.RepositoryMeta) {
	perPage := 100
	for {
		commits, err := s.gc.GetCommits(
			repo.OwnerName, repo.Name, repo.Since,
			repo.LastPage, perPage,
		)

		if err != nil {
			if err.Error() == "rate limited" {
				if repo.Index {

					dbRepo := &repository.Repository{
						ID:              repo.ID,
						OwnerName:       repo.OwnerName,
						Name:            repo.Name,
						Description:     repo.Description,
						Language:        repo.Language,
						URL:             repo.URL,
						ForksCount:      repo.ForksCount,
						StarsCount:      repo.StarsCount,
						OpenIssuesCount: repo.OpenIssuesCount,
						WatchersCount:   repo.WatchersCount,
						CreatedAt:       repo.CreatedAt,
						UpdatedAt:       repo.UpdatedAt,
						LastPage:        repo.LastPage,
						Index:           false,
					}
					s.rs.SaveRepository(dbRepo)
				}
				if err := s.ResumeIndexing(); err != nil {
					log.Printf("error indexing commit for repo %s  %s", repo.Name, err.Error())
				}
				// After resuming, refetch commits for the same page
				continue
			}
			log.Printf("error indexing commit for repo %s  %s", repo.Name, err.Error())
		}

		// Process fetched commits
		for _, commit := range commits {
			// Retrieve or create the author
			author, err := s.as.GetOrCreateAuthor(ctx, commit.Author.Name, commit.Author.Email)
			if err != nil {
				continue
			}

			// Save the new commit
			newCommit := &repository.Commit{
				RepositoryID: repo.ID,
				AuthorID:     author.ID,
				CommitHash:   commit.CommitHash,
				Message:      commit.Message,
				Date:         commit.Date,
				LastPage:     repo.LastPage,
			}

			if err := s.cs.CreateCommit(newCommit); err != nil {
				log.Printf("Failed to save commit %s: %v", commit.CommitHash, err)
			}

			log.Printf("Saved commit #%s for %s", commit.CommitHash, repo.Name)
		}

		// Break if the commit list is smaller than `perPage`, meaning no more commits
		if len(commits) < perPage {
			break
		}
		repo.LastPage++
	}

	dbRepo := &repository.Repository{
		ID:              repo.ID,
		OwnerName:       repo.OwnerName,
		Name:            repo.Name,
		Description:     repo.Description,
		Language:        repo.Language,
		URL:             repo.URL,
		ForksCount:      repo.ForksCount,
		StarsCount:      repo.StarsCount,
		OpenIssuesCount: repo.OpenIssuesCount,
		WatchersCount:   repo.WatchersCount,
		CreatedAt:       repo.CreatedAt,
		UpdatedAt:       repo.UpdatedAt,
		LastPage:        repo.LastPage,
	}
	s.rs.SaveRepository(dbRepo)
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
