package routes

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/core/service"
)

func NewAuthorRouter(router *http.ServeMux, svc *service.AuthorService) {
	router.HandleFunc("/authors/top", svc.GetTopAuthors)
}
