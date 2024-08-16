package data

import (
	"fmt"
	"log"
	"os"

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
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "indexer"),
		getEnv("DB_PORT", "5432"),
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
	var commits []Commit
	err := s.db.Joins("Repository").Where("repositories.name = ?", repoName).Find(&commits).Error
	return commits, err
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
