package repository

import (
	"time"

	"github.com/just-nibble/git-service/internal/domain"
)

type Repository struct {
	ID                uint   `gorm:"primaryKey"`
	OwnerName         string `gorm:"index"`
	Name              string `gorm:"uniqueIndex"`
	Description       string
	Language          string
	URL               string
	ForksCount        int
	StarsCount        int
	OpenIssuesCount   int
	WatchersCount     int
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Commits           []Commit
	Since             time.Time
	LastPage          int
	LastFetchedCommit string
	Index             bool
}

func (pr *Repository) ToDomain() *domain.RepositoryMeta {
	return &domain.RepositoryMeta{
		ID:                pr.ID,
		Name:              pr.Name,
		Description:       pr.Description,
		URL:               pr.URL,
		Language:          pr.Language,
		ForksCount:        pr.ForksCount,
		StarsCount:        pr.StarsCount,
		OpenIssuesCount:   pr.OpenIssuesCount,
		WatchersCount:     pr.WatchersCount,
		CreatedAt:         pr.CreatedAt,
		UpdatedAt:         pr.UpdatedAt,
		LastFetchedCommit: pr.LastFetchedCommit,
		Index:             pr.Index,
		LastPage:          pr.LastPage,
	}
}

func ToGormRepo(r *domain.RepositoryMeta) *Repository {
	return &Repository{
		ID:                r.ID,
		Name:              r.Name,
		Description:       r.Description,
		URL:               r.URL,
		Language:          r.Language,
		ForksCount:        r.ForksCount,
		StarsCount:        r.StarsCount,
		OpenIssuesCount:   r.OpenIssuesCount,
		WatchersCount:     r.WatchersCount,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
		LastFetchedCommit: r.LastFetchedCommit,
		Index:             r.Index,
		LastPage:          r.LastPage,
	}
}
