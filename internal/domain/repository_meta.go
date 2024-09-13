package domain

import (
	"time"

	"github.com/just-nibble/git-service/internal/http/dtos"
)

type RepositoryMeta struct {
	ID                uint
	OwnerName         string
	Name              string
	Description       string
	Language          string
	URL               string
	ForksCount        int
	StarsCount        int
	OpenIssuesCount   int
	WatchersCount     int
	LastPage          int
	LastFetchedCommit string
	Index             bool
	Since             time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (r RepositoryMeta) ToDto() dtos.RepositoryMeta {
	return dtos.RepositoryMeta{
		Name:            r.Name,
		Description:     r.Description,
		URL:             r.URL,
		Language:        r.Language,
		ForksCount:      r.ForksCount,
		StarsCount:      r.StarsCount,
		OpenIssuesCount: r.OpenIssuesCount,
		WatchersCount:   r.WatchersCount,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}
