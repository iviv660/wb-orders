package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (api *API) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderUID := chi.URLParam(r, "orderUID")

	order, err := api.orderService.Get(r.Context(), orderUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
