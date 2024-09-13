package usecases

import (
	"context"

	"github.com/just-nibble/git-service/internal/http/dtos"
	"github.com/just-nibble/git-service/internal/repository"
)

type AuthorUseCase interface {
	GetTopAuthors(ctx context.Context, repoName string, limit int) ([]dtos.Author, error)
}

type authorUseCase struct {
	authorStore repository.AuthorStore
}

func NewAuthorUseCase(authorStore repository.AuthorStore) AuthorUseCase {
	return &authorUseCase{
		authorStore: authorStore,
	}
}

func (s *authorUseCase) GetTopAuthors(ctx context.Context, repoName string, limit int) ([]dtos.Author, error) {

	as, err := s.authorStore.GetTopAuthors(ctx, repoName, limit)
	if err != nil {
		return []dtos.Author{}, nil
	}

	var authors []dtos.Author

	for _, v := range as {
		author := dtos.Author{
			Name:        v.Name,
			Email:       v.Name,
			CommitCount: v.CommitCount,
		}

		authors = append(authors, author)
	}

	return authors, nil
}
