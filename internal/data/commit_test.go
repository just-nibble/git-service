package data

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateCommit(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	author := &Author{
		Name:  "author-name",
		Email: "author@example.com",
	}
	dbInstance := &GormRepositoryStore{db: db}
	err = dbInstance.db.Create(author).Error
	assert.NoError(t, err)

	commit := &Commit{
		CommitHash:   "abc123",
		AuthorID:     author.ID,
		RepositoryID: 1,
		Message:      "Initial commit",
		Date:         time.Now(),
	}

	err = dbInstance.CreateCommit(commit)
	assert.NoError(t, err)

	var fetchedCommit Commit
	err = dbInstance.db.First(&fetchedCommit, "commit_hash = ?", commit.CommitHash).Error
	assert.NoError(t, err)
	assert.Equal(t, commit.CommitHash, fetchedCommit.CommitHash)
	assert.Equal(t, commit.AuthorID, fetchedCommit.AuthorID)
	assert.Equal(t, commit.RepositoryID, fetchedCommit.RepositoryID)
	assert.Equal(t, commit.Message, fetchedCommit.Message)
}

func TestDuplicateCommit(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	author := &Author{
		Name:  "author-name",
		Email: "author@example.com",
	}
	dbInstance := &GormRepositoryStore{db: db}
	err = dbInstance.db.Create(author).Error
	assert.NoError(t, err)

	commit := &Commit{
		CommitHash:   "abc123",
		AuthorID:     author.ID,
		RepositoryID: 1,
		Message:      "Initial commit",
		Date:         time.Now(),
	}

	err = dbInstance.CreateCommit(commit)
	assert.NoError(t, err)

	err = dbInstance.CreateCommit(commit)
	assert.Error(t, err) // Expect error due to duplicate commit hash
}
