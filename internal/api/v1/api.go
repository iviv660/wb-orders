package v1

import (
	"app/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	orderService service.Service
	Router       http.Handler
}

func NewAPI(orderService service.Service) *API {
	api := &API{orderService: orderService}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/order/{orderUID}", api.GetOrder)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./index.html")
	})

	api.Router = r
	return api
}
