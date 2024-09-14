package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/just-nibble/git-service/internal/http/dtos"
	"github.com/just-nibble/git-service/internal/usecases"
	"github.com/just-nibble/git-service/pkg/errcodes"
	"github.com/just-nibble/git-service/pkg/response"
)

type RepositoryHandler struct {
	gitRepositoryUsecase usecases.RepoMetaUsecase
}

func NewRepositoryHandler(gitRepositoryUsecase usecases.RepoMetaUsecase) *RepositoryHandler {
	return &RepositoryHandler{
		gitRepositoryUsecase: gitRepositoryUsecase,
	}
}

func (rh RepositoryHandler) AddRepository(w http.ResponseWriter, r *http.Request) {
	var req dtos.RepositoryInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()

	_, err := rh.gitRepositoryUsecase.InitiateIndexing(ctx, req)
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
	repos, err := rh.gitRepositoryUsecase.RetrieveAllRepos(r.Context())
	if err != nil {
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(repos) == 0 {
		response.SuccessResponse(w, http.StatusOK, "no repository indexed yet")
		return
	}

	var repoResponse []dtos.RepositoryMeta

	for _, v := range repos {
		repo := dtos.RepositoryMeta{
			Name:        v.Name,
			URL:         v.URL,
			Description: v.Description,
			Language:    v.Language,
			Owner: struct {
				Login string "json:\"login\""
			}{
				Login: v.OwnerName,
			},
			ForksCount:      v.ForksCount,
			StarsCount:      v.StarsCount,
			OpenIssuesCount: v.OpenIssuesCount,
			WatchersCount:   v.WatchersCount,
			CreatedAt:       v.CreatedAt,
			UpdatedAt:       v.UpdatedAt,
		}
		repoResponse = append(repoResponse, repo)
	}

	response.SuccessResponse(w, http.StatusOK, repoResponse)
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

	repo, err := rh.gitRepositoryUsecase.FindRepoByName(ctx, repoName)
	if err != nil {
		if err == errcodes.ErrNoRecordFound {
			response.ErrorResponse(w, http.StatusBadRequest, "no repository found")
			return
		}
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	repoReponse := dtos.RepositoryMeta{
		Name:        repo.Name,
		URL:         repo.URL,
		Description: repo.Description,
		Language:    repo.Language,
		Owner: struct {
			Login string "json:\"login\""
		}{
			Login: repo.OwnerName,
		},
		ForksCount:      repo.ForksCount,
		StarsCount:      repo.StarsCount,
		OpenIssuesCount: repo.OpenIssuesCount,
		WatchersCount:   repo.WatchersCount,
		CreatedAt:       repo.CreatedAt,
		UpdatedAt:       repo.UpdatedAt,
	}
	response.SuccessResponse(w, http.StatusOK, repoReponse)
}
