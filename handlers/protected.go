package handlers

import "net/http"

func ProtectedRoute(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	w.Write([]byte("Welcome to load balancer , " + username))
}
