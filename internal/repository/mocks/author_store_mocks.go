package mocks

import (
	"context"

	"github.com/just-nibble/git-service/internal/repository"
	"github.com/stretchr/testify/mock"
)

// Mock implementation for the AuthorRepository interface
type AuthorRepository struct {
	mock.Mock
}

func (m *AuthorRepository) GetTopAuthors(ctx context.Context, repoName string, limit int) ([]repository.Author, error) {
	args := m.Called(ctx, repoName, limit)
	return args.Get(0).([]repository.Author), args.Error(1)
}
