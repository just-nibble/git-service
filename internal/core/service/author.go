package service

import (
	"net/http"
	"strconv"

	"github.com/just-nibble/git-service/internal/adapters/api"
	"github.com/just-nibble/git-service/internal/adapters/db"
	"github.com/just-nibble/git-service/pkg/response"
)

type AuthorService struct {
	as           db.AuthorStore
	githubClient *api.GitHubClient
}

func NewAuthorService(as db.AuthorStore, gc *api.GitHubClient) *AuthorService {
	return &AuthorService{
		as:           as,
		githubClient: gc,
	}
}

func (s *AuthorService) GetTopAuthors(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repo")
	if repoName == "" {
		http.Error(w, "Repository name is required", http.StatusBadRequest)
		return
	}

	nStr := r.URL.Query().Get("n")
	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		http.Error(w, "Invalid number of authors", http.StatusBadRequest)
		return
	}

	// Fetch top commit authors from the dbbase
	authors, err := s.as.GetTopAuthors(n)
	if err != nil {
		http.Error(w, "Failed to retrieve authors", http.StatusInternalServerError)
		return
	}

	response.SuccessResponse(w, http.StatusOK, authors)
}
