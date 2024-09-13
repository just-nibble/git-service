package repository

import (
	"context"

	"github.com/just-nibble/git-service/internal/domain"
)

// CommitRepository defines an interface for database operations
type CommitRepository interface {
	SaveCommit(ctx context.Context, commit domain.Commit) (*domain.Commit, error)
	GetCommitByHash(ctx context.Context, commitHash string) (*domain.Commit, error)
	GetCommitsByRepository(ctx context.Context, repoMetadata domain.RepositoryMeta, query domain.APIPaging) ([]domain.Commit, error)
}
