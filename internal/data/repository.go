package data

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// RepositoryStore defines an interface for database operations
type RepositoryStore interface {
	CreateRepository(repo *Repository) error
	GetRepositoryByName(name string) (*Repository, error)
	SaveCommit(commit *Commit) error
	GetCommitsByRepository(repoName string) ([]Commit, error)
	GetTopAuthors(limit int) ([]Author, error)
	GetAllRepositories() ([]Repository, error)
	GetCommitByHash(hash string) (*Commit, error)
	CreateCommit(commit *Commit) error
	GetOrCreateAuthor(name, email string) (*Author, error)
	ResetRepositoryStartDate(repoName string, since time.Time) error
}

// GormRepositoryStore is a GORM-based implementation of RepositoryStore
type GormRepositoryStore struct {
	db *gorm.DB
}

// NewGormRepositoryStore initializes a new GormRepositoryStore
func NewGormRepositoryStore(db *gorm.DB) *GormRepositoryStore {
	return &GormRepositoryStore{db: db}
}

func InitDB() *gorm.DB {
	// Set up PostgreSQL connection details
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Automatically migrate the schema
	if err := db.AutoMigrate(&Repository{}, &Commit{}, &Author{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// Helper function to fetch environment variables with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Implement the methods for RepositoryStore interface

func (s *GormRepositoryStore) CreateRepository(repo *Repository) error {
	return s.db.Create(repo).Error
}

func (s *GormRepositoryStore) GetRepositoryByName(name string) (*Repository, error) {
	var repo Repository
	err := s.db.Where("name = ?", name).First(&repo).Error
	return &repo, err
}

func (s *GormRepositoryStore) SaveCommit(commit *Commit) error {
	return s.db.Create(commit).Error
}

func (s *GormRepositoryStore) GetCommitsByRepository(repoName string) ([]Commit, error) {
	// var commits []Commit
	var repository Repository
	err := s.db.Preload("Commits.Author").Where("name = ?", repoName).Find(&repository).Error
	// err := s.db.Joins("Repository").Where("repositories.name = ?", repoName).Find(&commits).Error
	return repository.Commits, err
}

func (s *GormRepositoryStore) GetTopAuthors(limit int) ([]Author, error) {
	var authors []Author
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
func (s *GormRepositoryStore) GetAllRepositories() ([]Repository, error) {
	var repositories []Repository
	if err := s.db.Find(&repositories).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve repositories: %v", err)
	}
	return repositories, nil
}

// GetCommitByHash retrieves a commit by its hash
func (s *GormRepositoryStore) GetCommitByHash(hash string) (*Commit, error) {
	var commit Commit
	if err := s.db.Where("commit_hash = ?", hash).First(&commit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("commit not found")
		}
		return nil, fmt.Errorf("failed to retrieve commit: %v", err)
	}
	return &commit, nil
}

// CreateCommit inserts a new commit into the database
func (s *GormRepositoryStore) CreateCommit(commit *Commit) error {
	if err := s.db.Create(commit).Error; err != nil {
		return fmt.Errorf("failed to create commit: %v", err)
	}
	return nil
}

// GetOrCreateAuthor retrieves an existing author by name and email, or creates a new one if it does not exist.
func (s *GormRepositoryStore) GetOrCreateAuthor(name, email string) (*Author, error) {
	var author Author
	err := s.db.Where("name = ? AND email = ?", name, email).First(&author).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Author not found, create a new one
			author = Author{
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
	var repo Repository
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
