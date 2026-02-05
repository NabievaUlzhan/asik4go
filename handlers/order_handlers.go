package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"foodstore/models"
	"foodstore/storage"
)

type OrderHandler struct {
	Store *storage.Store
}

type CreateOrderRequest struct {
	UserID int `json:"user_id"`
	Items  []struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	} `json:"items"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.UserID <= 0 {
		writeError(w, http.StatusBadRequest, "user_id must be > 0")
		return
	}

	items := make([]models.OrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, models.OrderItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		})
	}

	order, createdItems, err := h.Store.CreateOrder(req.UserID, items)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"order": order,
		"items": createdItems,
	})
}

func (h *OrderHandler) UserOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 || parts[0] != "users" || parts[2] != "orders" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	userID, err := strconv.Atoi(parts[1])
	if err != nil || userID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	orders, err := h.Store.GetUserOrders(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, orders)
}
