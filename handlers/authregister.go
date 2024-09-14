package handlers

import (
	"auth/models"
	"auth/utils"
	"encoding/json"
	"net/http"
	"regexp"
	"sync"
	"unicode"
)

var (
	users        = make(map[string]models.User)
	usersMutex   sync.Mutex
	emailRegex   = regexp.MustCompile(`^[a-z0-9._+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
	psymbolRegex = regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`)
)

func AuthRegister(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate username's length
	if len(user.Username) < 3 || len(user.Username) >32 {
		http.Error(w, "Username must be at least 3 characters long", http.StatusBadRequest)
		return
	}

	// Validate email format
	if !emailRegex.MatchString(user.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Validate password 
	if len(user.Password) < 8 || len(user.Password) > 32 {
		http.Error(w, "Password must be between 8-32 characters long", http.StatusBadRequest)
		return
	}

	if !containsCapitalLetter(user.Password) {
		http.Error(w, "Password must contain at least one capital letter", http.StatusBadRequest)
		return
	}

	if !containsSymbol(user.Password) {
		http.Error(w, "Password must contain at least one symbol", http.StatusBadRequest)
		return
	}

	usersMutex.Lock()
	defer usersMutex.Unlock()

	// Check if user already exists
	if _, exists := users[user.Username]; exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Save user with hashed password
	user.Password = hashedPassword
	users[user.Username] = user

	

	w.WriteHeader(http.StatusCreated)
}

func containsCapitalLetter(password string) bool {
	for _, ch := range password {
		if unicode.IsUpper(ch) {
			return true
		}
	}
	return false
}

func containsSymbol(password string) bool {
	return psymbolRegex.MatchString(password)
}


