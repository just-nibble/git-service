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

type RepoMetaUsecase interface {
	InitiateIndexing(ctx context.Context, input dtos.RepositoryInput) (*domain.RepositoryMeta, error)
	FindRepoByName(ctx context.Context, name string) (*domain.RepositoryMeta, error)
	RetrieveAllRepos(ctx context.Context) ([]domain.RepositoryMeta, error)
	ResumeIndexing(ctx context.Context) error
	ModifyRepoStatus(ctx context.Context, active bool) error
}
type repoMetaUsecase struct {
	repoMetaRepo repository.RepositoryMetaRepository
	commitRepo   repository.CommitRepository
	authorRepo   repository.AuthorRepository
	gitClient    git.GitClient
	cfg          config.Config
	logger       log.Log
}

func NewrepoMetaUsecase(repoMetaRepo repository.RepositoryMetaRepository, commitRepo repository.CommitRepository, authorRepo repository.AuthorRepository, gitClient git.GitClient, cfg config.Config, logger log.Log) *repoMetaUsecase {
	return &repoMetaUsecase{
		repoMetaRepo: repoMetaRepo,
		commitRepo:   commitRepo,
		authorRepo:   authorRepo,
		gitClient:    gitClient,
		cfg:          cfg,
		logger:       logger,
	}
}

func (uc *repoMetaUsecase) FindRepoByName(ctx context.Context, name string) (*domain.RepositoryMeta, error) {
	repo, err := uc.repoMetaRepo.RepoMeta(ctx, name)
	if err != nil {
		uc.logger.Error.Printf("Could not find repository named %s: %s", name, err.Error())
		return nil, err
	}
	uc.logger.Info.Printf("Successfully found repository named %s", name)
	return repo, nil
}

func (uc *repoMetaUsecase) RetrieveAllRepos(ctx context.Context) ([]domain.RepositoryMeta, error) {
	repos, err := uc.repoMetaRepo.AllRepoMeta(ctx)
	if err != nil {
		uc.logger.Error.Printf("Failed to list all repositories: %s", err.Error())
		return nil, err
	}
	uc.logger.Info.Println("Successfully listed all repositories")
	return repos, nil
}

func (uc *repoMetaUsecase) ModifyRepoStatus(ctx context.Context, active bool) error {
	if err := uc.repoMetaRepo.UpdateRepositoryStatus(ctx, active); err != nil {
		uc.logger.Error.Printf("Failed to update repository status to %v: %s", active, err.Error())
		return err
	}
	uc.logger.Info.Printf("Repository status successfully updated to %v", active)
	return nil
}

func (uc *repoMetaUsecase) InitiateIndexing(ctx context.Context, input dtos.RepositoryInput) (*domain.RepositoryMeta, error) {
	if !validator.IsRepository(input.Name) {
		uc.logger.Error.Printf("Invalid format for repository name: %s", input.Name)
		return nil, errcodes.ErrInvalidRepositoryName
	}

	existingRepo, err := uc.repoMetaRepo.RepoMeta(ctx, input.Name)
	if err != nil && err != errcodes.ErrNoRecordFound {
		uc.logger.Error.Printf("Error while checking existence of repository %s: %s", input.Name, err.Error())
		return nil, err
	}

	if existingRepo != nil && existingRepo.Name != "" {
		uc.logger.Error.Printf("Repository %s is already added", input.Name)
		return nil, errcodes.ErrRepoAlreadyAdded
	}

	repoMeta, err := uc.gitClient.FetchRepoMetadata(ctx, input.Name)
	if err != nil {
		uc.logger.Error.Printf("Error fetching metadata for %s: %s", input.Name, err.Error())
		return nil, err
	}

	repoMeta.Index = true

	savedRepoMeta, err := uc.repoMetaRepo.SaveRepoMetadata(ctx, *repoMeta)
	if err != nil {
		uc.logger.Error.Printf("Failed to save metadata for repository %s: %s", input.Name, err.Error())
		return nil, err
	}

	uc.logger.Info.Printf("Indexing initiated for repository %s", input.Name)
	go uc.processIndexing(ctx, *savedRepoMeta)

	return savedRepoMeta, nil
}

func (uc *repoMetaUsecase) processIndexing(ctx context.Context, repo domain.RepositoryMeta) {
	page := repo.LastPage
	var latestCommit string

	uc.logger.Info.Printf("Starting commit retrieval for repository %s from page %d", repo.Name, page)
	for {
		select {
		case <-ctx.Done():
			uc.logger.Info.Printf("Indexing operation canceled for repository %s", repo.Name)
			return
		default:
			commits, hasMore, err := uc.gitClient.FetchCommits(ctx, repo, uc.cfg.DefaultStartDate, uc.cfg.DefaultEndDate, "", int(page), uc.cfg.GitCommitFetchPerPage)
			if err != nil {
				uc.logger.Error.Printf("Error retrieving commits for repository %s: %s", repo.Name, err.Error())
				time.Sleep(5 * time.Second)
				continue
			}

			for _, commit := range commits {
				commit.RepoID = repo.ID
				if _, err = uc.commitRepo.SaveCommit(ctx, commit); err != nil {
					uc.logger.Error.Printf("Error saving commit %s for repository %s: %s", commit.Hash, repo.Name, err.Error())
					continue
				}
				latestCommit = commit.Hash
			}

			repo.LastFetchedCommit = latestCommit
			repo.LastPage = page
			if _, err = uc.repoMetaRepo.UpdateRepoMetadata(ctx, repo); err != nil {
				uc.logger.Error.Printf("Failed to update metadata for repository %s: %v", repo.Name, err)
				continue
			}

			if !hasMore {
				repo.Index = false
				if _, err = uc.repoMetaRepo.UpdateRepoMetadata(ctx, repo); err != nil {
					uc.logger.Error.Printf("Error updating indexing status for repository %s: %s", repo.Name, err.Error())
				}
				uc.logger.Info.Printf("Indexing finished for repository %s", repo.Name)
				break
			}
			page++
		}
	}
}

func (uc *repoMetaUsecase) ResumeIndexing(ctx context.Context) error {
	uc.logger.Info.Println("Resuming indexing operations...")
	repositories, err := uc.repoMetaRepo.AllRepoMeta(ctx)
	if err != nil {
		uc.logger.Error.Printf("Failed to retrieve repositories for indexing continuation: %s", err.Error())
		return err
	}

	for _, repo := range repositories {
		go uc.monitorCommits(ctx, repo)
	}
	return nil
}

func (uc *repoMetaUsecase) monitorCommits(ctx context.Context, repo domain.RepositoryMeta) error {
	ticker := time.NewTicker(uc.cfg.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			uc.logger.Info.Printf("Stopped commit monitoring for repository %s", repo.Name)
			return ctx.Err()
		case <-ticker.C:
			repoMeta, err := uc.repoMetaRepo.RepoMeta(ctx, repo.Name)
			if err != nil {
				uc.logger.Error.Printf("Error retrieving repository metadata %s: %s", repo.Name, err.Error())
				return err
			}

			if !repoMeta.Index {
				uc.logger.Info.Printf("Resuming commit fetching for repository %s", repo.Name)
				uc.updateCommits(ctx, *repoMeta)
			}
		}
	}
}

func (uc *repoMetaUsecase) updateCommits(ctx context.Context, repo domain.RepositoryMeta) {
	uc.logger.Info.Printf("Continuing commit reconciliation for repository %s", repo.Name)
	page := repo.LastPage
	lastCommit := repo.LastFetchedCommit
	endDate := uc.cfg.DefaultEndDate

	for {
		select {
		case <-ctx.Done():
			uc.logger.Info.Printf("Commit reconciliation halted for repository %s", repo.Name)
			return
		default:
			commits, hasMore, err := uc.gitClient.FetchCommits(ctx, repo, uc.cfg.DefaultStartDate, endDate, lastCommit, int(page), uc.cfg.GitCommitFetchPerPage)
			if err != nil {
				uc.logger.Error.Printf("Error fetching commits for repository %s: %s", repo.Name, err.Error())
				return
			}

			if len(commits) == 0 {
				uc.logger.Info.Printf("No new commits for repository %s, resetting page to 1", repo.Name)
				page = 1
				lastCommit = ""
				continue
			}

			for _, commit := range commits {
				if _, err = uc.commitRepo.GetCommitByHash(ctx, commit.Hash); err != nil {
					if err == errcodes.ErrNoRecordFound {
						commit.RepoID = repo.ID
						if _, err = uc.commitRepo.SaveCommit(ctx, commit); err != nil {
							uc.logger.Error.Printf("Error saving commit %s for repository %s: %s", commit.Hash, repo.Name, err.Error())
							continue
						}
						lastCommit = commit.Hash
					} else {
						uc.logger.Error.Printf("Error retrieving commit %s for repository %s: %s", commit.Hash, repo.Name, err.Error())
					}
				}
			}

			repo.LastFetchedCommit = lastCommit
			repo.LastPage = page
			if _, err = uc.repoMetaRepo.UpdateRepoMetadata(ctx, repo); err != nil && err != errcodes.ErrContextCancelled {
				uc.logger.Error.Printf("Error updating repository metadata %s: %v", repo.Name, err)
				return
			}

			if !hasMore {
				uc.logger.Info.Printf("No more commits to fetch for repository %s", repo.Name)
				break
			}
			page++
			endDate = time.Now()
		}
	}
}
