package usecases

import (
	"context"
	"testing"

	"github.com/just-nibble/git-service/internal/repository"
	"github.com/just-nibble/git-service/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestAuthorUseCase_GetTopAuthors_Success tests the success scenario for GetTopAuthors
func TestAuthorUseCase_GetTopAuthors(t *testing.T) {
	// Arrange
	mockAuthorRepository := new(mocks.AuthorRepository)
	mockAuthors := []repository.Author{
		{ID: 1, Name: "John Doe", CommitCount: 10},
		{ID: 2, Name: "Jane Smith", CommitCount: 5},
	}

	mockAuthorRepository.On("GetTopAuthors", mock.Anything, "repo1", 2).Return(mockAuthors, nil)

	uc := NewAuthorUseCase(mockAuthorRepository)

	// Act
	authors, err := uc.GetTopAuthors(context.TODO(), "repo1", 2)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(authors))
	assert.Equal(t, "John Doe", authors[0].Name)
	assert.Equal(t, "Jane Smith", authors[1].Name)
	mockAuthorRepository.AssertExpectations(t)
}
