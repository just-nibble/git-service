package service

import (
	"context"

	"github.com/just-nibble/git-service/internal/adapters/api"
	"github.com/just-nibble/git-service/internal/adapters/repository"
	"github.com/just-nibble/git-service/internal/core/domain/entities"
)

type AuthorService interface {
	GetTopAuthors(ctx context.Context, limit int) ([]entities.Author, error)
}

type authorService struct {
	as           repository.AuthorStore
	githubClient *api.GitHubClient
}

func NewAuthorService(as repository.AuthorStore, gc *api.GitHubClient) *authorService {
	return &authorService{
		as:           as,
		githubClient: gc,
	}
}

func (s *authorService) GetTopAuthors(ctx context.Context, limit int) ([]entities.Author, error) {

	as, err := s.as.GetTopAuthors(ctx, limit)
	if err != nil {
		return []entities.Author{}, nil
	}

	var authors []entities.Author

	for _, v := range as {
		author := entities.Author{
			ID:          v.ID,
			Name:        v.Name,
			Email:       v.Name,
			CommitCount: v.CommitCount,
		}

		authors = append(authors, author)
	}

	return authors, nil
}
