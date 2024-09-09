package db

import (
	"fmt"

	"github.com/just-nibble/git-service/internal/core/domain/entities"
	"gorm.io/gorm"
)

// AuthorStore defines an interface for database operations
type AuthorStore interface {
	GetTopAuthors(limit int) ([]entities.Author, error)
	GetOrCreateAuthor(name, email string) (*entities.Author, error)
}

// GormAuthorStore is a GORM-based implementation of AuthorStore
type GormAuthorStore struct {
	db *gorm.DB
}

// NewGormAuthorStore initializes a new GormAuthorStore
func NewGormAuthorStore(db *gorm.DB) *GormAuthorStore {
	return &GormAuthorStore{db: db}
}

func (s *GormAuthorStore) GetTopAuthors(limit int) ([]entities.Author, error) {
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

// GetOrCreateAuthor retrieves an existing author by name and email, or creates a new one if it does not exist.
func (s *GormAuthorStore) GetOrCreateAuthor(name, email string) (*entities.Author, error) {
	var author entities.Author
	err := s.db.Where("name = ? AND email = ?", name, email).Limit(1).Find(&author).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve author: %v", err)
	}

	if author.ID == 0 {
		// Author not found, create a new one
		author = entities.Author{
			Name:  name,
			Email: email,
		}
		if err := s.db.Create(&author).Error; err != nil {
			return nil, fmt.Errorf("failed to create author: %v", err)
		}
	}

	return &author, nil
}
