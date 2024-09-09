package db

import (
	"fmt"

	"github.com/just-nibble/git-service/internal/core/domain/entities"
	"gorm.io/gorm"
)

// CommitStore defines an interface for database operations
type CommitStore interface {
	SaveCommit(commit *entities.Commit) error
	GetCommitsByRepository(repoName string) ([]entities.Commit, error)
	GetCommitByHash(hash string) (*entities.Commit, error)
	CreateCommit(commit *entities.Commit) error
}

// GormCommitStore is a GORM-based implementation of CommitStore
type GormCommitStore struct {
	db *gorm.DB
}

// NewGormCommitStore initializes a new GormCommitStore
func NewGormCommitStore(db *gorm.DB) *GormCommitStore {
	return &GormCommitStore{db: db}
}

func (s *GormCommitStore) SaveCommit(commit *entities.Commit) error {
	return s.db.Create(commit).Error
}

func (s *GormCommitStore) GetCommitsByRepository(repoName string) ([]entities.Commit, error) {
	// var commits []Commit
	var repository entities.Repository
	err := s.db.Preload("Commits.Author").Where("name = ?", repoName).Find(&repository).Error
	// err := s.db.Joins("Repository").Where("repositories.name = ?", repoName).Find(&commits).Error
	return repository.Commits, err
}

// GetCommitByHash retrieves a commit by its hash
func (s *GormCommitStore) GetCommitByHash(hash string) (*entities.Commit, error) {
	var commit entities.Commit
	err := s.db.Where("commit_hash = ?", hash).Limit(1).Find(&commit).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve commit: %v", err)
	}
	return &commit, nil
}

// CreateCommit inserts a new commit into the database
func (s *GormCommitStore) CreateCommit(commit *entities.Commit) error {
	if err := s.db.Create(commit).Error; err != nil {
		return fmt.Errorf("failed to create commit: %v", err)
	}
	return nil
}
