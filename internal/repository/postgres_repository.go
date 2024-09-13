package repository

import (
	"context"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/just-nibble/git-service/pkg/errcodes"
	"gorm.io/gorm"
)

// GormRepositoryStore is a GORM-based implementation of RepositoryStore
type GormRepositoryStore struct {
	db *gorm.DB
}

// NewGormRepositoryStore initializes a new GormRepositoryStore
func NewGormRepositoryStore(db *gorm.DB) RepositoryStore {
	return &GormRepositoryStore{db: db}
}

// Implement the methods for RepositoryStore interface

func (r *GormRepositoryStore) SaveRepoMetadata(ctx context.Context, repo domain.RepositoryMeta) (*domain.RepositoryMeta, error) {
	dbRepository := ToGormRepo(&repo)

	err := r.db.WithContext(ctx).Create(dbRepository).Error
	if err != nil {
		return nil, err
	}
	return dbRepository.ToDomain(), err
}

func (r *GormRepositoryStore) RepoMetadataByPublicId(ctx context.Context, publicId string) (*domain.RepositoryMeta, error) {
	if ctx.Err() == context.Canceled {
		return nil, errcodes.ErrContextCancelled
	}

	var repo Repository
	err := r.db.WithContext(ctx).Where("public_id = ?", publicId).Find(&repo).Error

	if repo.ID == 0 {
		return nil, errcodes.ErrNoRecordFound
	}
	return repo.ToDomain(), err
}

func (r *GormRepositoryStore) RepoMetadataByName(ctx context.Context, name string) (*domain.RepositoryMeta, error) {
	if ctx.Err() == context.Canceled {
		return nil, errcodes.ErrContextCancelled
	}
	var repo Repository
	err := r.db.WithContext(ctx).Where("name = ?", name).Find(&repo).Error
	if repo.ID == 0 {
		return nil, errcodes.ErrNoRecordFound
	}
	return repo.ToDomain(), err
}

func (r *GormRepositoryStore) AllRepoMetadata(ctx context.Context) ([]domain.RepositoryMeta, error) {
	var dbRepositories []Repository

	err := r.db.WithContext(ctx).Find(&dbRepositories).Error

	if err != nil {
		return nil, err
	}

	var repoMetaDataResponse []domain.RepositoryMeta

	for _, dbRepository := range dbRepositories {
		repoMetaDataResponse = append(repoMetaDataResponse, *dbRepository.ToDomain())
	}
	return repoMetaDataResponse, err
}

func (r *GormRepositoryStore) UpdateRepoMetadata(ctx context.Context, repo domain.RepositoryMeta) (*domain.RepositoryMeta, error) {
	if ctx.Err() == context.Canceled {
		return nil, errcodes.ErrContextCancelled
	}
	dbRepo := ToGormRepo(&repo)

	err := r.db.WithContext(ctx).Model(&Repository{}).Where(&Repository{ID: repo.ID}).Updates(&dbRepo).Error
	if err != nil {
		return nil, err
	}

	return dbRepo.ToDomain(), nil
}

func (r *GormRepositoryStore) UpdateFetchingStateForAllRepos(ctx context.Context, isFetching bool) error {
	return r.db.WithContext(ctx).Model(&Repository{}).
		Where("index = ?", true).
		Update("index", isFetching).
		Error
}
