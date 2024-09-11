package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/just-nibble/git-service/internal/adapters/http/dtos"
	"github.com/just-nibble/git-service/internal/core/service"
	"github.com/just-nibble/git-service/pkg/response"
)

type RepositoryHandler struct {
	repoService   service.RepositoryService
	commitService service.CommitService
}

func NewRepositoryHandler(repoService service.RepositoryService, commitService service.CommitService) *RepositoryHandler {
	return &RepositoryHandler{repoService: repoService, commitService: commitService}
}

func (h *RepositoryHandler) AddRepository(w http.ResponseWriter, r *http.Request) {
	var req dtos.RepositoryInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Parse the 'since' date
	sinceDate, err := time.Parse("2006-01-02", req.Since)
	if err != nil {
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid date format for 'since'")
		return
	}

	if req.Owner == "" || req.Name == "" {
		response.ErrorResponse(w, http.StatusBadRequest, "Owner and repository name are required")
		return
	}

	// Fetch repository details from GitHub
	_, err = h.repoService.CreateRepository(req.Owner, req.Name, sinceDate)
	if err != nil {
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	repo, err := h.repoService.GetRepositoryByName(req.Name)

	go func() { h.commitService.IndexCommits(ctx, repo) }()

	response.SuccessResponse(w, http.StatusCreated, repo)
}

func (h *RepositoryHandler) ResetStartDate(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repo")

	var req struct {
		Since string `json:"since"` // Expecting an ISO date string
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Parse the provided date
	since, err := time.Parse(time.RFC3339, req.Since)
	if err != nil {
		http.Error(w, "Invalid date format. Use RFC3339 format", http.StatusBadRequest)
		return
	}

	// Call the service to reset the start date
	if err := h.repoService.ResetRepositoryStartDate(repoName, since); err != nil {
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessResponse(w, http.StatusCreated, "Start date reset successfully")
}
