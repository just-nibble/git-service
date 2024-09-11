package handlers

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/core/service"
	"github.com/just-nibble/git-service/pkg/response"
)

type CommitHandler struct {
	service service.CommitService
}

func NewCommitHandler(service service.CommitService) *CommitHandler {
	return &CommitHandler{service: service}
}

func (h *CommitHandler) GetCommitsByRepo(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repo")
	if repoName == "" {
		http.Error(w, "Repository name is required", http.StatusBadRequest)
		return
	}

	// Fetch commits from the dbbase
	commits, err := h.service.GetCommitsByRepo(repoName)
	if err != nil {
		http.Error(w, "Failed to retrieve commits", http.StatusInternalServerError)
		return
	}

	response.SuccessResponse(w, http.StatusOK, commits)
}
