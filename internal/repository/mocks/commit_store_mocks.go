package mocks

import (
	"context"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/just-nibble/git-service/internal/http/dtos"
	"github.com/stretchr/testify/mock"
)

// CommitStore mock
type CommitStore struct {
	mock.Mock
}

func (m *CommitStore) GetCommitsByRepository(ctx context.Context, repoMetadata domain.RepositoryMeta, query dtos.APIPagingDto) (*dtos.MultiCommitsResponse, error) {
	args := m.Called(ctx, repoMetadata, query)
	return args.Get(0).(*dtos.MultiCommitsResponse), args.Error(1)
}

func (m *CommitStore) SaveCommit(ctx context.Context, commit domain.Commit) (*domain.Commit, error) {
	args := m.Called(ctx, commit)
	return args.Get(0).(*domain.Commit), args.Error(1)
}

func (m *CommitStore) GetCommitByHash(ctx context.Context, commitHash string) (*domain.Commit, error) {
	args := m.Called(ctx, commitHash)
	return args.Get(0).(*domain.Commit), args.Error(1)
}
