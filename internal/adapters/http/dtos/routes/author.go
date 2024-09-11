package routes

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/adapters/http/handlers"
)

func NewAuthorRouter(router *http.ServeMux, handler handlers.AuthorHandler) {
	router.HandleFunc("/authors/top", handler.GetTopAuthors)
}
