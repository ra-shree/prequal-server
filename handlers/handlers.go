package handlers

import (
	"auth/middleware"
	"auth/utils"
	"log"
	"net/http"
)

func Handler() {
	//dns for users db
	
	if err := utils.InitDB(); err != nil {
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
