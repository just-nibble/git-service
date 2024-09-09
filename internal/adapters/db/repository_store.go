package db

import (
	"fmt"
	"time"

	"github.com/just-nibble/git-service/internal/core/domain/entities"
	"gorm.io/gorm"
)

// RepositoryStore defines an interface for database operations
type RepositoryStore interface {
	CreateRepository(repo *entities.Repository) error
	GetRepositoryByName(name string) (*entities.Repository, error)
	GetAllRepositories() ([]entities.Repository, error)
	ResetRepositoryStartDate(repoName string, since time.Time) error
	CountRepository() (int64, error)
}

// GormRepositoryStore is a GORM-based implementation of RepositoryStore
type GormRepositoryStore struct {
	db *gorm.DB
}

// NewGormRepositoryStore initializes a new GormRepositoryStore
func NewGormRepositoryStore(db *gorm.DB) *GormRepositoryStore {
	return &GormRepositoryStore{db: db}
}

// Implement the methods for RepositoryStore interface

func (s *GormRepositoryStore) CreateRepository(repo *entities.Repository) error {
	return s.db.Create(repo).Error
}

func (s *GormRepositoryStore) GetRepositoryByName(name string) (*entities.Repository, error) {
	var repo entities.Repository
	err := s.db.Where("name = ?", name).Limit(1).Find(&repo).Error
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

// GetAllRepositories retrieves all repositories from the database
func (s *GormRepositoryStore) GetAllRepositories() ([]entities.Repository, error) {
	var repositories []entities.Repository
	if err := s.db.Find(&repositories).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve repositories: %v", err)
	}
	return repositories, nil
}

func (s *GormRepositoryStore) ResetRepositoryStartDate(repoName string, since time.Time) error {
	var repo entities.Repository
	if err := s.db.Where("name = ?", repoName).First(&repo).Error; err != nil {
		return err
	}

	// Update the since date
	repo.Since = since
	if err := s.db.Save(&repo).Error; err != nil {
		return err
	}

	return nil
}

func (s *GormRepositoryStore) CountRepository() (int64, error) {
	var count int64
	if err := s.db.Model(&entities.Repository{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
