package usecases

import (
	"context"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/just-nibble/git-service/internal/repository"
)

type AuthorUseCase interface {
	GetTopAuthors(ctx context.Context, repoName string, limit int) ([]domain.Author, error)
}

type authorUseCase struct {
	authorRepository repository.AuthorRepository
}

func NewAuthorUseCase(authorRepository repository.AuthorRepository) AuthorUseCase {
	return &authorUseCase{
		authorRepository: authorRepository,
	}
}

func (s *authorUseCase) GetTopAuthors(ctx context.Context, repoName string, limit int) ([]domain.Author, error) {

	as, err := s.authorRepository.GetTopAuthors(ctx, repoName, limit)
	if err != nil {
		return []domain.Author{}, nil
	}

	var authors []domain.Author

	for _, v := range as {
		author := domain.Author{
			Name:        v.Name,
			Email:       v.Name,
			CommitCount: v.CommitCount,
		}

		authors = append(authors, author)
	}

	return authors, nil
}
