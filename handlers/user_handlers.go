package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"foodstore/storage"
)

type UserHandler struct {
	Store *storage.Store
}

func (h *UserHandler) Users(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	users := h.Store.GetAllUsers()

	resp := make([]map[string]any, 0, len(users))
	for _, u := range users {
		resp = append(resp, map[string]any{
			"id":    u.ID,
			"name":  u.Name,
			"email": u.Email,
			"role":  u.Role,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) UserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	user, ok := h.Store.GetUserByID(id)
	if !ok {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	resp := map[string]any{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
