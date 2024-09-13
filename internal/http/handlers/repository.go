package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/just-nibble/git-service/internal/http/dtos"
	"github.com/just-nibble/git-service/internal/usecases"
	"github.com/just-nibble/git-service/pkg/errcodes"
	"github.com/just-nibble/git-service/pkg/response"
)

type RepositoryHandler struct {
	gitRepositoryUsecase usecases.GitRepositoryUsecase
}

func NewRepositoryHandler(gitRepositoryUsecase usecases.GitRepositoryUsecase) *RepositoryHandler {
	return &RepositoryHandler{
		gitRepositoryUsecase: gitRepositoryUsecase,
	}
}

func (rh RepositoryHandler) AddRepository(w http.ResponseWriter, r *http.Request) {
	var req dtos.RepositoryInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	_, err := rh.gitRepositoryUsecase.StartIndexing(ctx, req)
	if err != nil {
		if err == errcodes.ErrRepoAlreadyAdded {
			response.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessResponse(w, http.StatusCreated, "Repository successfully indexed, its commits are being fetched...")
}

func (rh RepositoryHandler) FetchAllRepositories(w http.ResponseWriter, r *http.Request) {
	repos, err := rh.gitRepositoryUsecase.GetAll(r.Context())
	if err != nil {
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(repos) == 0 {
		response.SuccessResponse(w, http.StatusOK, "no repository indexed yet")
		return
	}
	response.SuccessResponse(w, http.StatusOK, repos)
}

func (rh RepositoryHandler) FetchRepository(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	if owner == "" {
		response.ErrorResponse(w, http.StatusBadRequest, "Repository owner is required")
		return
	}

	name := r.PathValue("name")
	if name == "" {
		response.ErrorResponse(w, http.StatusBadRequest, "Repository name is required")
		return
	}

	repoName := fmt.Sprintf("%s/%s", owner, name)

	ctx := r.Context()

	repo, err := rh.gitRepositoryUsecase.GetByName(ctx, repoName)
	if err != nil {
		if err == errcodes.ErrNoRecordFound {
			response.ErrorResponse(w, http.StatusBadRequest, "no repository found")
			return
		}
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessResponse(w, http.StatusOK, repo)
}
