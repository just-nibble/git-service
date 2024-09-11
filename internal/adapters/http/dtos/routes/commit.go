package routes

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/adapters/http/handlers"
)

func NewCommitRouter(router *http.ServeMux, handler handlers.CommitHandler) {
	router.HandleFunc("/commits", handler.GetCommitsByRepo)
}
