package repository

import (
	"context"

	"github.com/just-nibble/git-service/internal/domain"
)

// RepositoryMetaRepository defines an interface for database operations
type RepositoryMetaRepository interface {
	SaveRepoMetadata(ctx context.Context, repository domain.RepositoryMeta) (*domain.RepositoryMeta, error)
	UpdateRepoMetadata(ctx context.Context, repo domain.RepositoryMeta) (*domain.RepositoryMeta, error)
	RepoMeta(ctx context.Context, name string) (*domain.RepositoryMeta, error)
	AllRepoMeta(ctx context.Context) ([]domain.RepositoryMeta, error)
	UpdateRepositoryStatus(ctx context.Context, isFetching bool) error
}
