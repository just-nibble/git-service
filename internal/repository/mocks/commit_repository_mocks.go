package mocks

import (
	"context"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// CommitRepository mock
type CommitRepository struct {
	mock.Mock
}

func (m *CommitRepository) GetCommitsByRepository(ctx context.Context, repoMetadata domain.RepositoryMeta, query domain.APIPaging) ([]domain.Commit, error) {
	args := m.Called(ctx, repoMetadata, query)
	return args.Get(0).([]domain.Commit), args.Error(1)
}

func (m *CommitRepository) SaveCommit(ctx context.Context, commit domain.Commit) (*domain.Commit, error) {
	args := m.Called(ctx, commit)
	return args.Get(0).(*domain.Commit), args.Error(1)
}

func (m *CommitRepository) GetCommitByHash(ctx context.Context, commitHash string) (*domain.Commit, error) {
	args := m.Called(ctx, commitHash)
	return args.Get(0).(*domain.Commit), args.Error(1)
}
