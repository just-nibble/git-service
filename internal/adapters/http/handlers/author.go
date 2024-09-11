package handlers

import (
	"net/http"
	"strconv"

	"github.com/just-nibble/git-service/internal/core/service"
	"github.com/just-nibble/git-service/pkg/response"
)

type AuthorHandler struct {
	service service.AuthorService
}

func NewAuthorHandler(service service.AuthorService) *AuthorHandler {
	return &AuthorHandler{service: service}
}

func (h *AuthorHandler) GetTopAuthors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	nStr := r.URL.Query().Get("n")
	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		http.Error(w, "Invalid number of authors", http.StatusBadRequest)
		return
	}

	authors, err := h.service.GetTopAuthors(ctx, n)
	if err != nil {
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessResponse(w, http.StatusOK, authors)
}
