package usecases

import (
	"context"
	"time"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/just-nibble/git-service/internal/http/dtos"
	"github.com/just-nibble/git-service/internal/repository"
	"github.com/just-nibble/git-service/pkg/config"
	"github.com/just-nibble/git-service/pkg/errcodes"
	"github.com/just-nibble/git-service/pkg/git"
	"github.com/just-nibble/git-service/pkg/log"
	"github.com/just-nibble/git-service/pkg/validator"
)

type GitRepositoryUsecase interface {
	StartIndexing(ctx context.Context, input dtos.RepositoryInput) (*domain.RepositoryMeta, error)
	GetById(ctx context.Context, repoId string) (*domain.RepositoryMeta, error)
	GetByName(ctx context.Context, repoName string) (*domain.RepositoryMeta, error)
	GetAll(ctx context.Context) ([]domain.RepositoryMeta, error)
	ResumeFetching(ctx context.Context) error
	UpdateFetchingStatusForAllRepositories(ctx context.Context, status bool) error
}

type gitRepoUsecase struct {
	repoMetaRepository repository.RepositoryMetaRepository
	commitRepository   repository.CommitRepository
	authorRepository   repository.AuthorRepository
	gitClient          git.GitClient
	config             config.Config
	log                log.Log
}

func NewGitRepositoryUsecase(repoMetaRepository repository.RepositoryMetaRepository, commitRepository repository.CommitRepository,
	authorRepository repository.AuthorRepository, gitClient git.GitClient, config config.Config, log log.Log) GitRepositoryUsecase {
	return &gitRepoUsecase{
		repoMetaRepository: repoMetaRepository,
		commitRepository:   commitRepository,
		authorRepository:   authorRepository,
		gitClient:          gitClient,
		config:             config,
		log:                log,
	}
}

func (uc *gitRepoUsecase) GetById(ctx context.Context, repoId string) (*domain.RepositoryMeta, error) {
	repo, err := uc.repoMetaRepository.RepoMetadataByPublicId(ctx, repoId)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (uc *gitRepoUsecase) GetByName(ctx context.Context, repoName string) (*domain.RepositoryMeta, error) {
	repo, err := uc.repoMetaRepository.RepoMetadataByName(ctx, repoName)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (uc *gitRepoUsecase) GetAll(ctx context.Context) ([]domain.RepositoryMeta, error) {
	repos, err := uc.repoMetaRepository.AllRepoMetadata(ctx)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func (uc *gitRepoUsecase) StartIndexing(ctx context.Context, input dtos.RepositoryInput) (*domain.RepositoryMeta, error) {
	//validate repository name to ensure it has owner and repo name
	if !validator.IsRepository(input.Name) {
		return nil, errcodes.ErrInvalidRepositoryName
	}

	// ensure repo does not exist on the db
	repo, err := uc.repoMetaRepository.RepoMetadataByName(ctx, input.Name)
	if err != nil && err != errcodes.ErrNoRecordFound {
		return nil, err
	}

	if repo != nil && repo.Name != "" {
		return nil, errcodes.ErrRepoAlreadyAdded
	}

	repoMetadata, err := uc.gitClient.FetchRepoMetadata(ctx, input.Name)
	if err != nil {
		return nil, err
	}

	repoMetadata.CreatedAt = time.Now()
	repoMetadata.UpdatedAt = time.Now()
	repoMetadata.Index = true

	sRepoMetadata, err := uc.repoMetaRepository.SaveRepoMetadata(ctx, *repoMetadata)
	if err != nil {
		return nil, err
	}

	uc.log.Info.Println("indexing repo...")
	go uc.startRepoIndexing(ctx, *sRepoMetadata)

	return repo, nil
}

func (uc *gitRepoUsecase) startRepoIndexing(ctx context.Context, repo domain.RepositoryMeta) {
	page := repo.LastPage
	lastFetchedCommit := ""

	uc.log.Info.Printf("fetching commits for repo: %s, starting from page-%d", repo.Name, page)
	for {
		commits, morePages, err := uc.gitClient.FetchCommits(ctx, repo, uc.config.DefaultStartDate, uc.config.DefaultEndDate, "", int(page), uc.config.GitCommitFetchPerPage)
		if err != nil {
			uc.log.Error.Printf("Failed to fetch commits for repository %s: %s", repo.Name, err.Error())
			continue
		}

		// loop through commits and persist each
		for _, commit := range commits {

			commit.RepoID = repo.ID

			_, err = uc.commitRepository.SaveCommit(ctx, commit)
			if err != nil {
				uc.log.Error.Printf("error saving commit-id:%s for repo %s %s", commit.Hash, repo.Name, err.Error())
				continue
			}
			lastFetchedCommit = commit.Hash
		}

		// Update the repository's last fetched commit in the database
		repo.LastFetchedCommit = lastFetchedCommit
		repo.LastPage = page
		_, err = uc.repoMetaRepository.UpdateRepoMetadata(ctx, repo)
		if err != nil {
			uc.log.Info.Printf("Error updating repository %s: %v", repo.Name, err)
			continue
		}

		if !morePages {
			// update isFetching to false as flag for start of monitoring
			repo.Index = false
			_, err = uc.repoMetaRepository.UpdateRepoMetadata(ctx, repo)
			if err != nil {
				uc.log.Error.Printf("Error updating isFetching column of repository %s: %s", repo.Name, err.Error())
			}
			break
		}
		page++
	}
}

func (uc *gitRepoUsecase) ResumeFetching(ctx context.Context) error {
	uc.log.Info.Println("Resume fetching started...")
	repos, err := uc.repoMetaRepository.AllRepoMetadata(ctx)
	if err != nil {
		uc.log.Error.Printf("Error fetching repositories from database: %s", err.Error())
		return err
	}
	uc.log.Info.Printf("Saved repos %v", repos)
	for _, repo := range repos {
		go uc.startPeriodicFetching(ctx, repo)
	}
	return nil
}

func (uc *gitRepoUsecase) startPeriodicFetching(ctx context.Context, repo domain.RepositoryMeta) error {
	ticker := time.NewTicker(uc.config.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			uc.log.Info.Printf("Git repository [%s] commits monitoring service stopped", repo.Name)
			return ctx.Err()
		case <-ticker.C:
			r, err := uc.repoMetaRepository.RepoMetadataByName(ctx, repo.Name)
			if err != nil {
				uc.log.Info.Printf("error getting repo metadata for monitoring: %s", err.Error())
				return err
			}
			if !r.Index {
				uc.log.Info.Printf("Commits periodic fetching started for repo %v", repo.Name)
				uc.fetchAndReconcileCommits(ctx, *r)
			}
		}
	}
}

func (uc *gitRepoUsecase) fetchAndReconcileCommits(ctx context.Context, repo domain.RepositoryMeta) {
	uc.log.Info.Printf("Resume fetching and reconciling commits for repo: %s", repo.Name)
	page := repo.LastPage

	lastFetchedCommit := repo.LastFetchedCommit

	until := uc.config.DefaultEndDate

	for {
		select {
		case <-ctx.Done():
			uc.log.Info.Printf("Git repository [%s] fetchAndReconcileCommits service stopped", repo.Name)
			return
		default:
			commits, morePages, err := uc.gitClient.FetchCommits(ctx, repo, uc.config.DefaultStartDate, until, lastFetchedCommit, int(page), uc.config.GitCommitFetchPerPage)
			if err != nil {
				uc.log.Error.Printf("Error fetching commits for repo %s: %s", repo.Name, err.Error())
				return
			}

			if len(commits) == 0 {
				uc.log.Info.Printf("No new commits for repo %s, resetting page to 1", repo.Name)
				page = 1               //reset the page
				lastFetchedCommit = "" //don't use sha endpoint
				continue
			}

			for _, commit := range commits {
				_, err = uc.commitRepository.GetCommitByHash(ctx, commit.Hash)
				if err != nil && err != errcodes.ErrNoRecordFound && err != errcodes.ErrContextCancelled {
					uc.log.Info.Printf("error getting commit by commit-hash:%s", commit.Hash)
				}
				if err == errcodes.ErrNoRecordFound {
					commit.RepoID = repo.ID

					_, err = uc.commitRepository.SaveCommit(ctx, commit)
					if err != nil {
						uc.log.Info.Printf("error saving commit-id:%s for repo %s", commit.Hash, repo.Name)
						continue
					}
					lastFetchedCommit = commit.Hash
				}
			}

			repo.LastFetchedCommit = lastFetchedCommit
			repo.LastPage = page
			_, err = uc.repoMetaRepository.UpdateRepoMetadata(ctx, repo)
			if err != nil && err != errcodes.ErrContextCancelled {
				uc.log.Info.Printf("Error updating repository %s: %v", repo.Name, err)
				return
			}

			if !morePages {
				uc.log.Info.Printf("no more page to fech for repo: %s", repo.Name)
				break
			}

			page++

			until = time.Now()
		}
	}
}

func (uc *gitRepoUsecase) UpdateFetchingStatusForAllRepositories(ctx context.Context, status bool) error {
	return uc.repoMetaRepository.UpdateFetchingStateForAllRepos(ctx, status)
}
