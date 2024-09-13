package repository

import (
	"context"
)

// AuthorStore defines an interface for database operations
type AuthorStore interface {
	GetTopAuthors(ctx context.Context, repoName string, limit int) ([]Author, error)
}
