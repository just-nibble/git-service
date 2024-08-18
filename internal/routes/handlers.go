package routes

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/service"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(svc *service.IndexerService) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/repositories", svc.AddRepository)
	router.HandleFunc("/commits/", svc.GetCommitsByRepo)
	router.HandleFunc("/authors/top", svc.GetTopAuthors)
	// Serve Swagger documentation
	router.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	return router
}
