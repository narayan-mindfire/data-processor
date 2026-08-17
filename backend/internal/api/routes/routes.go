package routes

import (
	"net/http"

	"github.com/narayan-mindfire/data-processor/backend/internal/api/repositories"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/narayan-mindfire/data-processor/backend/docs"
)

func RegisterRoutes(db repositories.JobRepository) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api-docs/", httpSwagger.WrapHandler)

	RegisterJobRoutes(mux, "/api/v1/pipelines", db)

	return mux
}
