package repository

import (
	"fmt"
	"time"

	"github.com/just-nibble/git-service/internal/core/domain/entities"
	"gorm.io/gorm"
)

// Repository represents a GitHub repository
type Repository struct {
	ID              uint   `gorm:"primaryKey"`
	OwnerName       string `gorm:"index"`
	Name            string `gorm:"uniqueIndex"`
	Description     string
	Language        string
	URL             string
	ForksCount      int
	StarsCount      int
	OpenIssuesCount int
	WatchersCount   int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Commits         []Commit
	Since           time.Time
	LastPage        int
	Index           bool
}

// RepositoryStore defines an interface for database operations
type RepositoryStore interface {
	SaveRepository(repo *Repository) error
	GetRepositoryByName(name string) (*entities.RepositoryMeta, error)
	GetAllRepositories() ([]entities.RepositoryMeta, error)
	ResetRepositoryStartDate(repoName string, since time.Time) error
	CountRepository() (int64, error)
}

// GormRepositoryStore is a GORM-based implementation of RepositoryStore
type GormRepositoryStore struct {
	db *gorm.DB
}

// NewGormRepositoryStore initializes a new GormRepositoryStore
func NewGormRepositoryStore(db *gorm.DB) *GormRepositoryStore {
	return &GormRepositoryStore{db: db}
}

// Implement the methods for RepositoryStore interface

func (s *GormRepositoryStore) SaveRepository(repo *Repository) error {
	return s.db.Save(repo).Error
}

func (s *GormRepositoryStore) GetRepositoryByName(name string) (*entities.RepositoryMeta, error) {
	var repo Repository
	err := s.db.Where("name = ?", name).Limit(1).Find(&repo).Error
	if err != nil {
		return nil, err
	}

	repoMeta := entities.RepositoryMeta{
		ID:              repo.ID,
		OwnerName:       repo.OwnerName,
		Name:            repo.Name,
		Description:     repo.Description,
		Language:        repo.Language,
		URL:             repo.URL,
		ForksCount:      repo.ForksCount,
		StarsCount:      repo.StarsCount,
		OpenIssuesCount: repo.OpenIssuesCount,
		WatchersCount:   repo.WatchersCount,
		CreatedAt:       repo.CreatedAt,
		UpdatedAt:       repo.UpdatedAt,
		Since:           repo.Since,
	}
	return &repoMeta, nil
}

// GetAllRepositories retrieves all repositories from the database
func (s *GormRepositoryStore) GetAllRepositories() ([]entities.RepositoryMeta, error) {
	var repositories []Repository
	if err := s.db.Find(&repositories).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve repositories: %w", err)
	}

	var repoMetas []entities.RepositoryMeta

	for _, repo := range repositories {
		repoMeta := entities.RepositoryMeta{
			ID:              repo.ID,
			OwnerName:       repo.OwnerName,
			Name:            repo.Name,
			Description:     repo.Description,
			Language:        repo.Language,
			URL:             repo.URL,
			ForksCount:      repo.ForksCount,
			StarsCount:      repo.StarsCount,
			OpenIssuesCount: repo.OpenIssuesCount,
			WatchersCount:   repo.WatchersCount,
			CreatedAt:       repo.CreatedAt,
			UpdatedAt:       repo.UpdatedAt,
			Since:           repo.Since,
		}

		repoMetas = append(repoMetas, repoMeta)
	}

	return repoMetas, nil
}

func (s *GormRepositoryStore) ResetRepositoryStartDate(repoName string, since time.Time) error {
	var repo Repository
	if err := s.db.Where("name = ?", repoName).First(&repo).Error; err != nil {
		return err
	}

	// Update the since date
	repo.Since = since
	if err := s.db.Save(&repo).Error; err != nil {
		return err
	}

	return nil
}

func (s *GormRepositoryStore) CountRepository() (int64, error) {
	var count int64
	if err := s.db.Model(&Repository{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
