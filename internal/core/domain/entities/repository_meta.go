package entities

import (
	"time"
)

type RepositoryMeta struct {
	ID              uint
	OwnerName       string
	Name            string
	Description     string
	Language        string
	URL             string
	ForksCount      int
	StarsCount      int
	OpenIssuesCount int
	WatchersCount   int
	LastPage        int
	Index           bool
	Since           time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
