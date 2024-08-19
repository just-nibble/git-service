package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupDB() (*gorm.DB, error) {
	dsn := "host=db user=postgres password=mysecretpassword dbname=testdb port=5432 sslmode=disable" // Adjust this DSN as needed
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Repository{}); err != nil {
		return nil, err
	}
	return db, nil
}

func TestCreateRepository(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	repo := &Repository{
		ID:        1,
		Name:      "test-repo",
		URL:       "http://example.com",
		OwnerName: "test-owner",
	}

	dbInstance := &GormRepositoryStore{db: db}
	err = dbInstance.db.Create(repo).Error
	assert.NoError(t, err)

	var fetchedRepo Repository
	err = dbInstance.db.First(&fetchedRepo, repo.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, repo.ID, fetchedRepo.ID)
	assert.Equal(t, repo.Name, fetchedRepo.Name)
	assert.Equal(t, repo.URL, fetchedRepo.URL)
	assert.Equal(t, repo.OwnerName, fetchedRepo.OwnerName)
}
