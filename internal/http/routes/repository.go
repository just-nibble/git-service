package routes

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/http/handlers"
)

func NewRepositoryRouter(router *http.ServeMux, handler handlers.RepositoryHandler) {
	router.HandleFunc("/repositories", handler.AddRepository)
	router.HandleFunc("/repositories/{owner}/{name}", handler.FetchRepository)
}
