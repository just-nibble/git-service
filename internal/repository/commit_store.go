package repository

import (
	"context"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/just-nibble/git-service/internal/http/dtos"
)

// CommitStore defines an interface for database operations
type CommitStore interface {
	SaveCommit(ctx context.Context, commit domain.Commit) (*domain.Commit, error)
	GetCommitByHash(ctx context.Context, commitHash string) (*domain.Commit, error)
	GetCommitsByRepository(ctx context.Context, repoMetadata domain.RepositoryMeta, query dtos.APIPagingDto) (*dtos.MultiCommitsResponse, error)
}
