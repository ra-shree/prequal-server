package handlers

import (
	"log"
	"net/http"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {

    log.Println("GetUsers handler called") 
    
    // users, err := models.GetUsers()
    // if err != nil {
    //     http.Error(w, "Failed to get users", http.StatusInternalServerError)
    //     log.Printf("Error retrieving users: %v", err)
    //     return
    // }

    // // Encode the users as JSON and send them in the response
    // w.Header().Set("Content-Type", "application/json")
    // if err := json.NewEncoder(w).Encode(users); err != nil {
    //     http.Error(w, "Failed to encode users", http.StatusInternalServerError)
    //     return
    // }
}
