package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/just-nibble/git-service/internal/http/dtos"
	"github.com/just-nibble/git-service/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestGitCommitUsecase_GetAllCommitsByRepository_Success tests the success scenario
func TestGitCommitUsecase_GetAllCommitsByRepository_Success(t *testing.T) {
	// Arrange
	mockCommitStore := new(mocks.CommitStore)
	mockRepoStore := new(mocks.RepositoryStore)

	mockRepoMeta := &domain.RepositoryMeta{ID: 1, Name: "repo1"} // Use the correct type here (domain.RepositoryMeta)

	commitTime := time.Now()

	mockCommitsResp := &dtos.MultiCommitsResponse{
		Commits: []dtos.Commit{
			{
				SHA: "123",
				Commit: struct {
					Message string      `json:"message"`
					Author  dtos.Author `json:"author"`
					URL     string      `json:"url"`
				}{
					Message: "Initial commit",
					Author: dtos.Author{
						Name:  "John Doe",
						Email: "john.doe@example.com",
						Date:  commitTime,
					},
					URL: "http://example.com/123",
				},
			},
			{
				SHA: "456",
				Commit: struct {
					Message string      `json:"message"`
					Author  dtos.Author `json:"author"`
					URL     string      `json:"url"`
				}{
					Message: "Second commit",
					Author: dtos.Author{
						Name:  "Jane Smith",
						Email: "jane.smith@example.com",
						Date:  commitTime,
					},
					URL: "http://example.com/456",
				},
			},
		},
	}

	query := dtos.APIPagingDto{Page: 1, Limit: 10}

	// Update the mock to return domain.RepositoryMeta
	mockRepoStore.On("RepoMetadataByName", mock.Anything, "repo1").Return(mockRepoMeta, nil)
	mockCommitStore.On("GetCommitsByRepository", mock.Anything, *mockRepoMeta, query).Return(mockCommitsResp, nil)

	uc := NewGitCommitUsecase(mockCommitStore, mockRepoStore)

	// Act
	commits, err := uc.GetAllCommitsByRepository(context.TODO(), "repo1", query)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(commits.Commits))
	assert.Equal(t, "123", commits.Commits[0].SHA)
	assert.Equal(t, "Initial commit", commits.Commits[0].Commit.Message)
	assert.Equal(t, "John Doe", commits.Commits[0].Commit.Author.Name)
	assert.Equal(t, "john.doe@example.com", commits.Commits[0].Commit.Author.Email)
	assert.Equal(t, commitTime, commits.Commits[0].Commit.Author.Date)
	assert.Equal(t, "http://example.com/123", commits.Commits[0].Commit.URL)

	assert.Equal(t, "456", commits.Commits[1].SHA)
	assert.Equal(t, "Second commit", commits.Commits[1].Commit.Message)
	assert.Equal(t, "Jane Smith", commits.Commits[1].Commit.Author.Name)
	assert.Equal(t, "jane.smith@example.com", commits.Commits[1].Commit.Author.Email)
	assert.Equal(t, commitTime, commits.Commits[1].Commit.Author.Date)
	assert.Equal(t, "http://example.com/456", commits.Commits[1].Commit.URL)

	mockRepoStore.AssertExpectations(t)
	mockCommitStore.AssertExpectations(t)
}
