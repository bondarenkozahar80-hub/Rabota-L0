package http

import (
	"encoding/json"
	"net/http"
	"order-service/internal/domain "
	"order-service/internal/service"

	"github.com/gorilla/mux"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) GetOrderByUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderUID := vars["id"]

	if orderUID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	order, err := h.orderService.GetOrderByUID(r.Context(), orderUID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) ServeStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/static/index.html")
}
