package api

import "net/http"

type OrderServer interface {
	GetOrder(w http.ResponseWriter, r *http.Request)
}
