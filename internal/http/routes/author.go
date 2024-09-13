package routes

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/http/handlers"
)

func NewAuthorRouter(router *http.ServeMux, handler handlers.AuthorHandler) {
	router.HandleFunc("/authors/{owner}/{name}/top", handler.GetTopAuthors)
}
