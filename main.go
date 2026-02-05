package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"foodstore/handlers"
	"foodstore/storage"
)

func main() {
	store := storage.NewStore()

	go orderWorker(store)

	ph := &handlers.ProductHandler{Store: store}
	oh := &handlers.OrderHandler{Store: store}
	uh := &handlers.UserHandler{Store: store}
	ah := &handlers.AuthHandler{Store: store}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"message":"Food Store API is running","version":"1.0"}`))
	})

	mux.HandleFunc("/auth/register", ah.Register)
	mux.HandleFunc("/auth/login", ah.Login)
	mux.HandleFunc("/auth/logout", ah.Logout)
	mux.HandleFunc("/me", ah.Me)

	mux.HandleFunc("/api/users", uh.Users)
	mux.HandleFunc("/api/users/", uh.UserByID)

	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.RequireAuth(store, ph.Products)(w, r)
			return
		}
		ph.Products(w, r)
	})

	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut || r.Method == http.MethodDelete {
			handlers.RequireAuth(store, ph.ProductByID)(w, r)
			return
		}
		ph.ProductByID(w, r)
	})

	mux.HandleFunc("/orders", handlers.RequireAuth(store, oh.CreateOrder))

	mux.HandleFunc("/users/", handlers.RequireAuth(store, oh.UserOrders))

	port := 8080
	log.Printf("Server started at http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), logMiddleware(mux)))
}

func orderWorker(store *storage.Store) {
	for orderID := range store.OrderQueue {
		time.Sleep(1 * time.Second)
		log.Printf("[WORKER] Order %d processed in background\n", orderID)
	}
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}
