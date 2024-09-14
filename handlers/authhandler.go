package handlers

import (
	"auth/middleware"
	"auth/models"
	"log"
	"net/http"
)

func AuthHandle() {
	
	dsn := "postgres://admin:Admin01!@localhost:5432/usertable?sslmode=disable"
	if err := models.InitDB(dsn); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Route setup
	http.HandleFunc("/register", AuthRegister)
	http.HandleFunc("/login", AuthLogin)
	http.Handle("/protected", middleware.AuthMiddleware(http.HandlerFunc(ProtectedRoute)))
	http.HandleFunc("/users", GetUsers) 

	// Start the server
	log.Println("Server is running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
