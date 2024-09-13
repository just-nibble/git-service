package repository

import (
	"context"

	"github.com/just-nibble/git-service/internal/domain"
)

// RepositoryStore defines an interface for database operations
type RepositoryStore interface {
	SaveRepoMetadata(ctx context.Context, repository domain.RepositoryMeta) (*domain.RepositoryMeta, error)
	UpdateRepoMetadata(ctx context.Context, repo domain.RepositoryMeta) (*domain.RepositoryMeta, error)
	RepoMetadataByPublicId(ctx context.Context, publicId string) (*domain.RepositoryMeta, error)
	RepoMetadataByName(ctx context.Context, name string) (*domain.RepositoryMeta, error)
	AllRepoMetadata(ctx context.Context) ([]domain.RepositoryMeta, error)
	UpdateFetchingStateForAllRepos(ctx context.Context, isFetching bool) error
}
