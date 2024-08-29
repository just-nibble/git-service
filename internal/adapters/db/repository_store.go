package db

import (
	"fmt"
	"os"
	"time"

	"github.com/just-nibble/git-service/internal/core/domain/entities"
	"gorm.io/gorm"
)

// RepositoryStore defines an interface for database operations
type RepositoryStore interface {
	CreateRepository(repo *entities.Repository) error
	GetRepositoryByName(name string) (*entities.Repository, error)
	SaveCommit(commit *entities.Commit) error
	GetCommitsByRepository(repoName string) ([]entities.Commit, error)
	GetTopAuthors(limit int) ([]entities.Author, error)
	GetAllRepositories() ([]entities.Repository, error)
	GetCommitByHash(hash string) (*entities.Commit, error)
	CreateCommit(commit *entities.Commit) error
	GetOrCreateAuthor(name, email string) (*entities.Author, error)
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

// Helper function to fetch environment variables with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Implement the methods for RepositoryStore interface

func (s *GormRepositoryStore) CreateRepository(repo *entities.Repository) error {
	return s.db.Create(repo).Error
}

func (s *GormRepositoryStore) GetRepositoryByName(name string) (*entities.Repository, error) {
	var repo entities.Repository
	err := s.db.Where("name = ?", name).First(&repo).Error
	return &repo, err
}

func (s *GormRepositoryStore) SaveCommit(commit *entities.Commit) error {
	return s.db.Create(commit).Error
}

func (s *GormRepositoryStore) GetCommitsByRepository(repoName string) ([]entities.Commit, error) {
	// var commits []Commit
	var repository entities.Repository
	err := s.db.Preload("Commits.Author").Where("name = ?", repoName).Find(&repository).Error
	// err := s.db.Joins("Repository").Where("repositories.name = ?", repoName).Find(&commits).Error
	return repository.Commits, err
}

func (s *GormRepositoryStore) GetTopAuthors(limit int) ([]entities.Author, error) {
	var authors []entities.Author
	err := s.db.Raw(`
		SELECT authors.id, authors.name, authors.email, COUNT(commits.id) as commit_count
		FROM authors
		JOIN commits ON commits.author_id = authors.id
		GROUP BY authors.id
		ORDER BY commit_count DESC
		LIMIT ?
	`, limit).Scan(&authors).Error
	return authors, err
}

// GetAllRepositories retrieves all repositories from the database
func (s *GormRepositoryStore) GetAllRepositories() ([]entities.Repository, error) {
	var repositories []entities.Repository
	if err := s.db.Find(&repositories).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve repositories: %v", err)
	}
	return repositories, nil
}

// GetCommitByHash retrieves a commit by its hash
func (s *GormRepositoryStore) GetCommitByHash(hash string) (*entities.Commit, error) {
	var commit entities.Commit
	if err := s.db.Where("commit_hash = ?", hash).First(&commit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("commit not found")
		}
		return nil, fmt.Errorf("failed to retrieve commit: %v", err)
	}
	return &commit, nil
}

// CreateCommit inserts a new commit into the database
func (s *GormRepositoryStore) CreateCommit(commit *entities.Commit) error {
	if err := s.db.Create(commit).Error; err != nil {
		return fmt.Errorf("failed to create commit: %v", err)
	}
	return nil
}

// GetOrCreateAuthor retrieves an existing author by name and email, or creates a new one if it does not exist.
func (s *GormRepositoryStore) GetOrCreateAuthor(name, email string) (*entities.Author, error) {
	var author entities.Author
	err := s.db.Where("name = ? AND email = ?", name, email).First(&author).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Author not found, create a new one
			author = entities.Author{
				Name:  name,
				Email: email,
			}
			if err := s.db.Create(&author).Error; err != nil {
				return nil, fmt.Errorf("failed to create author: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to retrieve author: %v", err)
		}
	}
	return &author, nil
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
