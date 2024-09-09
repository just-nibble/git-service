package routes

import (
	"net/http"

	"github.com/just-nibble/git-service/internal/core/service"
)

func NewRepositoryRouter(router *http.ServeMux, svc *service.RepositoryService) {
	router.HandleFunc("/repositories", svc.AddRepository)
}
