package repository

import (
	"time"

	"github.com/just-nibble/git-service/internal/domain"
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

func (c *Commit) ToDomain() *domain.Commit {
	author := domain.Author{
		Name:  c.Author.Name,
		Email: c.Author.Email,
	}

	return &domain.Commit{
		Hash:    c.CommitHash,
		Message: c.Message,
		Author:  author,
		Date:    c.Date,
	}
}

func ToGormCommit(c *domain.Commit) *Commit {

	return &Commit{
		CommitHash:   c.Hash,
		Message:      c.Message,
		AuthorID:     c.AuthorID,
		Date:         c.Date,
		RepositoryID: c.RepoID,
	}
}
