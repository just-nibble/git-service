package repository

import (
	"context"

	"gorm.io/gorm"
)

// GormAuthorRepository is a GORM-based implementation of AuthorRepository
type GormAuthorRepository struct {
	db *gorm.DB
}

// NewGormAuthorRepository initializes a new GormAuthorRepository
func NewGormAuthorRepository(db *gorm.DB) AuthorRepository {
	return &GormAuthorRepository{db: db}
}

func (s *GormAuthorRepository) GetTopAuthors(ctx context.Context, repoName string, limit int) ([]Author, error) {
	var authors []Author
	err := s.db.WithContext(ctx).
		Table("author").
		Select("author.id, author.name, author.email, COUNT(commit.id) as commit_count").
		Joins("JOIN commit ON commit.author_id = author.id").
		Joins("JOIN repository ON commit.repository_id = repository.id").
		Where("repository.name = ?", repoName).
		Group("author.id").
		Order("commit_count DESC").
		Limit(limit).
		Find(&authors).
		Error
	return authors, err
}
