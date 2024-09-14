package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/just-nibble/git-service/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestGitCommitUsecase_GetAllCommitsByRepository_Success tests the success scenario
func TestGitCommitUsecase_GetAllCommitsByRepository_Success(t *testing.T) {
	// Arrange
	mockCommitRepository := new(mocks.CommitRepository)
	mockRepoRepository := new(mocks.RepositoryRepository)

	mockRepoMeta := &domain.RepositoryMeta{ID: 1, Name: "repo1"} // Use the correct type here (domain.RepositoryMeta)

	commitTime := time.Now()

	mockCommitsResp := []domain.Commit{
		{
			Hash:    "123",
			Message: "Initial commit",
			Author: domain.Author{
				Name:  "John Doe",
				Email: "john.doe@example.com",
			},
			Date: commitTime,
		},
		{
			Hash:    "456",
			Message: "Second commit",
			Author: domain.Author{
				Name:  "Jane Smith",
				Email: "jane.smith@example.com",
			},
			Date: commitTime,
		},
	}

	query := domain.APIPaging{Page: 1, Limit: 10}

	// Update the mock to return domain.RepositoryMeta
	mockRepoRepository.On("RepoMeta", mock.Anything, "repo1").Return(mockRepoMeta, nil)
	mockCommitRepository.On("GetCommitsByRepository", mock.Anything, *mockRepoMeta, query).Return(mockCommitsResp, nil)

	uc := NewGitCommitUsecase(mockCommitRepository, mockRepoRepository)

	// Act
	commits, err := uc.GetAllCommitsByRepository(context.TODO(), "repo1", query)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(commits))
	assert.Equal(t, "123", commits[0].Hash)
	assert.Equal(t, "Initial commit", commits[0].Message)
	assert.Equal(t, "John Doe", commits[0].Author.Name)
	assert.Equal(t, "john.doe@example.com", commits[0].Author.Email)
	assert.Equal(t, commitTime, commits[0].Date)

	assert.Equal(t, "456", commits[1].Hash)
	assert.Equal(t, "Second commit", commits[1].Message)
	assert.Equal(t, "Jane Smith", commits[1].Author.Name)
	assert.Equal(t, "jane.smith@example.com", commits[1].Author.Email)
	assert.Equal(t, commitTime, commitTime, commits[1].Date)

	mockRepoRepository.AssertExpectations(t)
	mockCommitRepository.AssertExpectations(t)
}
