package entities

import (
	"time"
)

// Repository represents a GitHub repository
type Repository struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	OwnerName       string    `json:"owner_name" gorm:"index"`
	Name            string    `json:"repo_name" gorm:"uniqueIndex"`
	Description     string    `json:"description"`
	Language        string    `json:"language"`
	URL             string    `json:"url"`
	ForksCount      int       `json:"forks_count"`
	StarsCount      int       `json:"stargazers_count"`
	OpenIssuesCount int       `json:"open_issues_count"`
	WatchersCount   int       `json:"watchers_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Commits         []Commit  `json:"commits" gorm:"foreignKey:RepositoryID"`
	Since           time.Time `json:"since"`
}

// Commit represents a commit in a repository
type Commit struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	CommitHash   string    `json:"hash" gorm:"uniqueIndex"`
	AuthorID     uint      `json:"author_id"`
	RepositoryID uint      `json:"repo_id"`
	Message      string    `json:"message"`
	Date         time.Time `json:"date"`
	Author       Author    `json:"author" gorm:"foreignKey:AuthorID"`
	CreatedAt    time.Time `json:"created_at"`
}

// Author represents the author of a commit
type Author struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"index"`
	Email       string    `json:"email" gorm:"index"`
	CommitCount int       `json:"commit_count"`
	Commits     []Commit  `json:"-" gorm:"foreignKey:AuthorID"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}
