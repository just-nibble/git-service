package mocks

import (
	"context"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// RepositoryRepository mock
type RepositoryRepository struct {
	mock.Mock
}

func (m *RepositoryRepository) RepoMeta(ctx context.Context, repoName string) (*domain.RepositoryMeta, error) {
	args := m.Called(ctx, repoName)
	return args.Get(0).(*domain.RepositoryMeta), args.Error(1)
}

func (m *RepositoryRepository) SaveRepoMetadata(ctx context.Context, repository domain.RepositoryMeta) (*domain.RepositoryMeta, error) {
	args := m.Called(ctx, repository)
	return args.Get(0).(*domain.RepositoryMeta), args.Error(1)
}
func (m *RepositoryRepository) UpdateRepoMetadata(ctx context.Context, repo domain.RepositoryMeta) (*domain.RepositoryMeta, error) {
	args := m.Called(ctx, repo)
	return args.Get(0).(*domain.RepositoryMeta), args.Error(1)
}

func (m *RepositoryRepository) RepoMetadataByPublicId(ctx context.Context, publicId string) (*domain.RepositoryMeta, error) {
	args := m.Called(ctx, publicId)
	return args.Get(0).(*domain.RepositoryMeta), args.Error(1)
}

func (m *RepositoryRepository) AllRepoMeta(ctx context.Context) ([]domain.RepositoryMeta, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.RepositoryMeta), args.Error(1)
}

func (m *RepositoryRepository) UpdateRepositoryStatus(ctx context.Context, isFetching bool) error {
	args := m.Called(ctx, isFetching)
	return args.Error(1)
}
