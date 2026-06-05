package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type NotificationRequest struct {
	Event   string `json:"event"`
	Message string `json:"message"`
}

type NotificationResponse struct {
	Status string `json:"status"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthResponse{Status: "UP"})
	})

	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req NotificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Log notification event simulation
		log.Printf("[Notification] RECEIVED EVENT: %s - MESSAGE: %s", req.Event, req.Message)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(NotificationResponse{Status: "dispatched"})
	})

	log.Printf("[Notification] Starting Notification Service on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start Notification Service: %v", err)
	}
}
