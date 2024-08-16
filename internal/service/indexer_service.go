package service

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/data"
)

type IndexerService struct {
	db data.RepositoryStore
}

func NewIndexerService(db data.RepositoryStore) *IndexerService {
	return &IndexerService{db: db}
}

func (s *IndexerService) AddRepository(w http.ResponseWriter, r *http.Request) {
	// Logic for adding a repository
}

func (s *IndexerService) GetCommitsByRepo(w http.ResponseWriter, r *http.Request) {
	// Logic for retrieving commits by repository
}

func (s *IndexerService) GetTopAuthors(w http.ResponseWriter, r *http.Request) {
	// Logic for getting top N authors
}
