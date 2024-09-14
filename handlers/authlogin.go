package handlers

import (
	"auth/models"
	"auth/utils"
	"encoding/json"
	"net/http"
)



func AuthLogin(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	

	usersMutex.Lock()
	defer usersMutex.Unlock()

	// Fetch user from DB
	user, exists := users[creds.Username]
	if !exists || !utils.CheckPasswordHash(creds.Password, user.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	tokenString, err := utils.GenerateJWT(creds.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Set JWT as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
		//Secure:   true, // Enable in production if using HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	w.Write([]byte("Login successful"))
}