package data

import (
	"time"
)

// Repository represents a GitHub repository
type Repository struct {
	ID        uint   `gorm:"primaryKey"`
	OwnerName string `gorm:"index"`
	Name      string `gorm:"uniqueIndex"`
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
	Commits   []Commit  `gorm:"foreignKey:RepositoryID"`
	Since     time.Time `json:"since"`
}

// Commit represents a commit in a repository
type Commit struct {
	ID           uint   `gorm:"primaryKey"`
	CommitHash   string `gorm:"uniqueIndex"`
	AuthorID     uint
	RepositoryID uint
	Message      string
	Date         time.Time
	Author       Author `gorm:"foreignKey:AuthorID"`
	CreatedAt    time.Time
}

// Author represents the author of a commit
type Author struct {
	ID        uint     `gorm:"primaryKey"`
	Name      string   `gorm:"index"`
	Email     string   `gorm:"index"`
	Commits   []Commit `gorm:"foreignKey:AuthorID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
