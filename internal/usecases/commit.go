package usecases

import (
	"context"

	"github.com/just-nibble/git-service/internal/http/dtos"
	"github.com/just-nibble/git-service/internal/repository"
)

type GitCommitUsecase interface {
	GetAllCommitsByRepository(ctx context.Context, repoName string, query dtos.APIPagingDto) (*dtos.MultiCommitsResponse, error)
}

type gitCommitUsecase struct {
	commitStore     repository.CommitStore
	repositoryStore repository.RepositoryStore
}

func NewGitCommitUsecase(commitStore repository.CommitStore, repositoryStore repository.RepositoryStore) GitCommitUsecase {
	return &gitCommitUsecase{
		commitStore:     commitStore,
		repositoryStore: repositoryStore,
	}
}

func (u *gitCommitUsecase) GetAllCommitsByRepository(ctx context.Context, repoName string, query dtos.APIPagingDto) (*dtos.MultiCommitsResponse, error) {
	// Fetch commits from the dbbase
	repoMetaData, err := u.repositoryStore.RepoMetadataByName(ctx, repoName)
	if err != nil {
		return nil, err
	}

	commitsResp, err := u.commitStore.GetCommitsByRepository(ctx, *repoMetaData, query)
	if err != nil {
		return nil, err
	}

	return commitsResp, nil
}
