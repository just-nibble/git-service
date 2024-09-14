package repository

import (
	"context"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/just-nibble/git-service/pkg/errcodes"
	"gorm.io/gorm"
)

// GormRepositoryRepository is a GORM-based implementation of RepositoryRepository
type GormRepositoryMetaRepository struct {
	db *gorm.DB
}

// NewGormRepositoryRepository initializes a new GormRepositoryRepository
func NewGormRepositoryMetaRepository(db *gorm.DB) RepositoryMetaRepository {
	return &GormRepositoryMetaRepository{db: db}
}

// Implement the methods for RepositoryMetaRepository interface

func (r *GormRepositoryMetaRepository) SaveRepoMetadata(ctx context.Context, repo domain.RepositoryMeta) (*domain.RepositoryMeta, error) {
	dbRepository := ToGormRepo(&repo)

	err := r.db.WithContext(ctx).Create(dbRepository).Error
	if err != nil {
		return nil, err
	}
	return dbRepository.ToDomain(), err
}

func (r *GormRepositoryMetaRepository) RepoMeta(ctx context.Context, name string) (*domain.RepositoryMeta, error) {
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

func (r *GormRepositoryMetaRepository) AllRepoMeta(ctx context.Context) ([]domain.RepositoryMeta, error) {
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

func (r *GormRepositoryMetaRepository) UpdateRepoMetadata(ctx context.Context, repo domain.RepositoryMeta) (*domain.RepositoryMeta, error) {
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

func (r *GormRepositoryMetaRepository) UpdateRepositoryStatus(ctx context.Context, isFetching bool) error {
	return r.db.WithContext(ctx).Model(&Repository{}).
		Where("index = ?", true).
		Update("index", isFetching).
		Error
}
