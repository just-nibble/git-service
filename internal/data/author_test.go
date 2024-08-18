package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOrCreateAuthor(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	dbInstance := &GormRepositoryStore{db: db}

	author, err := dbInstance.GetOrCreateAuthor("author-name", "author@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, author)

	// Verify that the author is created
	var fetchedAuthor Author
	err = dbInstance.db.First(&fetchedAuthor, author.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, author.Name, fetchedAuthor.Name)
	assert.Equal(t, author.Email, fetchedAuthor.Email)

	// Test retrieving the existing author
	existingAuthor, err := dbInstance.GetOrCreateAuthor("author-name", "author@example.com")
	assert.NoError(t, err)
	assert.Equal(t, author.ID, existingAuthor.ID)
}
