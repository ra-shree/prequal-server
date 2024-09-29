package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"
	"unicode"

	"github.com/ra-shree/prequal-server/pkg/common"
	"github.com/ra-shree/prequal-server/pkg/models"
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
	if len(user.Username) < 3 || len(user.Username) > 32 {
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
	hashedPassword, err := common.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Save user with hashed password
	user.Password = hashedPassword
	users[user.Username] = user

	// Insert the user into the database
	if err := models.InsertUser(user); err != nil {
		http.Error(w, "Error inserting user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
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
	if !exists || !common.CheckPasswordHash(creds.Password, user.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	tokenString, err := common.GenerateJWT(creds.Username)
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

func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetUsers handler called"))

	users, err := models.GetUsersinfo()
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func ProtectedRoute(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	w.Write([]byte("Welcome to load balancer , " + username))
}
