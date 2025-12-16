package v1

import (
	gen "app/internal/api/v1"
	"app/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Handler struct {
	orderService service.Service
}

func NewHandler(orderService service.Service) *Handler {
	return &Handler{orderService: orderService}
}

func NewAPI(svc service.Service) (http.Handler, error) {
	h := NewHandler(svc)

	ogenServer, err := gen.NewServer(h)
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		http.ServeFile(w, r, "api/openapi.yaml")
	})

	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!doctype html>
<html>
  <head><meta charset="utf-8"><title>API Docs</title></head>
  <body>
    <redoc spec-url="/openapi.yaml"></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
  </body>
</html>`))
	})

	r.Mount("/", ogenServer)

	return otelhttp.NewHandler(r, "http"), nil
}
