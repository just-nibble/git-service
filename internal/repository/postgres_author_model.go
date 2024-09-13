package repository

import "time"

type Author struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"index"`
	Email       string `gorm:"index"`
	CommitCount int
	Commits     []Commit `gorm:"foreignKey:AuthorID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
