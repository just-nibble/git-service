package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/just-nibble/git-service/internal/http/dtos"
	"github.com/just-nibble/git-service/internal/usecases"
	"github.com/just-nibble/git-service/pkg/response"
)

type CommitHandler struct {
	gitCommitUseCase usecases.GitCommitUsecase
}

func NewCommitHandler(gitCommitUseCase usecases.GitCommitUsecase) *CommitHandler {
	return &CommitHandler{gitCommitUseCase: gitCommitUseCase}
}

func getPagingInfo(r *http.Request) dtos.APIPagingDto {
	var paging dtos.APIPagingDto

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	sort := r.URL.Query().Get("sort")
	direction := r.URL.Query().Get("direction")

	paging.Limit = limit
	paging.Page = page
	paging.Sort = sort
	paging.Direction = direction

	return paging
}

func (h *CommitHandler) GetCommitsByRepoName(w http.ResponseWriter, r *http.Request) {
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

	query := getPagingInfo(r)

	// Fetch commits from the dbbase
	commits, err := h.gitCommitUseCase.GetAllCommitsByRepository(r.Context(), repoName, query)
	if err != nil {
		http.Error(w, "Failed to retrieve commits", http.StatusInternalServerError)
		return
	}

	response.SuccessResponse(w, http.StatusOK, commits)
}
