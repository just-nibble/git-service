package usecases

import (
	"context"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/just-nibble/git-service/internal/repository"
)

type GitCommitUsecase interface {
	GetAllCommitsByRepository(ctx context.Context, repoName string, query domain.APIPaging) ([]domain.Commit, error)
}

type gitCommitUsecase struct {
	commitRepository         repository.CommitRepository
	repositoryMetaRepository repository.RepositoryMetaRepository
}

func NewGitCommitUsecase(commitRepository repository.CommitRepository, repositoryRepository repository.RepositoryMetaRepository) GitCommitUsecase {
	return &gitCommitUsecase{
		commitRepository:         commitRepository,
		repositoryMetaRepository: repositoryRepository,
	}
}

func (u *gitCommitUsecase) GetAllCommitsByRepository(ctx context.Context, repoName string, query domain.APIPaging) ([]domain.Commit, error) {
	// Fetch commits from the dbbase
	repoMetaData, err := u.repositoryMetaRepository.RepoMeta(ctx, repoName)
	if err != nil {
		return nil, err
	}

	commitsResp, err := u.commitRepository.GetCommitsByRepository(ctx, *repoMetaData, query)
	if err != nil {
		return nil, err
	}

	return commitsResp, nil
}
