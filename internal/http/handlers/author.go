package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/just-nibble/git-service/internal/http/dtos"
	"github.com/just-nibble/git-service/internal/usecases"
	"github.com/just-nibble/git-service/pkg/response"
)

type AuthorHandler struct {
	authorUsecase usecases.AuthorUseCase
}

func NewAuthorHandler(authorUsecase usecases.AuthorUseCase) *AuthorHandler {
	return &AuthorHandler{authorUsecase: authorUsecase}
}

func (h *AuthorHandler) GetTopAuthors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	nStr := r.URL.Query().Get("n")
	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		http.Error(w, "Invalid number of authors", http.StatusBadRequest)
		return
	}

	authors, err := h.authorUsecase.GetTopAuthors(ctx, repoName, n)
	if err != nil {
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	var authorResponse []dtos.Author

	for _, v := range authors {
		author := dtos.Author{
			Name:        v.Name,
			Email:       v.Email,
			CommitCount: v.CommitCount,
		}
		authorResponse = append(authorResponse, author)
	}

	response.SuccessResponse(w, http.StatusOK, authorResponse)
}
