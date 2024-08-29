package routes

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/core/service"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(svc *service.Indexer) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/repositories", svc.AddRepository)
	router.HandleFunc("/commits/", svc.GetCommitsByRepo)
	router.HandleFunc("/authors/top", svc.GetTopAuthors)
	// Serve Swagger documentation
	router.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	return router
}
