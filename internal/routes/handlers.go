package routes

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/service"
)

type Router struct {
	service *service.IndexerService
}

func NewRouter(svc *service.IndexerService) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/repositories", svc.AddRepository)
	router.HandleFunc("/repositories/", svc.GetCommitsByRepo)
	router.HandleFunc("/authors/top", svc.GetTopAuthors)
	return router
}
