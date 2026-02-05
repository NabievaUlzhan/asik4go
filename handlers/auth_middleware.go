package handlers

import (
	"net/http"

	"foodstore/storage"
)

func GetUserIDFromRequest(store *storage.Store, r *http.Request) (int, bool) {
	c, err := r.Cookie("sid")
	if err != nil || c.Value == "" {
		return 0, false
	}
	return store.GetUserIDBySession(c.Value)
}

func RequireAuth(store *storage.Store, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUserIDFromRequest(store, r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "login required")
			return
		}
		next(w, r)
	}
}
