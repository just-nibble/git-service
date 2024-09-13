package repository

import (
	"context"
)

// AuthorRepository defines an interface for database operations
type AuthorRepository interface {
	GetTopAuthors(ctx context.Context, repoName string, limit int) ([]Author, error)
}
