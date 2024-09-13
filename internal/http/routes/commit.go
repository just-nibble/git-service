package routes

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/http/handlers"
)

func NewCommitRouter(router *http.ServeMux, handler handlers.CommitHandler) {
	router.HandleFunc("/commits/{owner}/{name}", handler.GetCommitsByRepoName)
}
