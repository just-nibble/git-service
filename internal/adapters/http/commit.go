package routes

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/core/service"
)

func NewCommitRouter(router *http.ServeMux, svc *service.CommitService) {
	router.HandleFunc("/commits/", svc.GetCommitsByRepo)
}
