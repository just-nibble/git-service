package repository

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Commit struct {
	ID           uint   `gorm:"primaryKey"`
	CommitHash   string `gorm:"uniqueIndex"`
	AuthorID     uint
	RepositoryID uint
	Message      string
	Date         time.Time
	Author       Author `gorm:"foreignKey:AuthorID"`
	CreatedAt    time.Time
	LastPage     int
}

// CommitStore defines an interface for database operations
type CommitStore interface {
	SaveCommit(commit *Commit) error
	GetCommitsByRepository(repoName string) ([]Commit, error)
	GetCommitByHash(hash string) (*Commit, error)
	CreateCommit(commit *Commit) error
}

// GormCommitStore is a GORM-based implementation of CommitStore
type GormCommitStore struct {
	db *gorm.DB
}

// NewGormCommitStore initializes a new GormCommitStore
func NewGormCommitStore(db *gorm.DB) *GormCommitStore {
	return &GormCommitStore{db: db}
}

func (s *GormCommitStore) SaveCommit(commit *Commit) error {
	return s.db.Save(commit).Error
}

func (s *GormCommitStore) GetCommitsByRepository(repoName string) ([]Commit, error) {
	// var commits []Commit
	var repository Repository
	err := s.db.Preload("Commits.Author").Where("name = ?", repoName).Find(&repository).Error
	// err := s.db.Joins("Repository").Where("repositories.name = ?", repoName).Find(&commits).Error
	return repository.Commits, err
}

// GetCommitByHash retrieves a commit by its hash
func (s *GormCommitStore) GetCommitByHash(hash string) (*Commit, error) {
	var commit Commit
	err := s.db.Where("commit_hash = ?", hash).Limit(1).Find(&commit).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve commit: %w", err)
	}
	return &commit, nil
}

// CreateCommit inserts a new commit into the database
func (s *GormCommitStore) CreateCommit(commit *Commit) error {
	if err := s.db.Create(commit).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil
		}
		return fmt.Errorf("failed to create commit: %w", err)
	}
	return nil
}
