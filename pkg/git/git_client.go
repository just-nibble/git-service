package git

import (
	"context"
	"time"

	"github.com/just-nibble/git-service/internal/domain"
)

type GitClient interface {
	FetchRepoMetadata(ctx context.Context, repositoryName string) (*domain.RepositoryMeta, error)
	FetchCommits(ctx context.Context, repo domain.RepositoryMeta, since time.Time, until time.Time, lastFetchedCommit string, page, perPage int) ([]domain.Commit, bool, error)
}
