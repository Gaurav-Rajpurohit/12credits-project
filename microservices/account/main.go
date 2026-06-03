package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

type AccountResponse struct {
	ID      string  `json:"id"`
	Balance float64 `json:"balance"`
	Status  string  `json:"status"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthResponse{Status: "UP"})
	})

	http.HandleFunc("/accounts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 || parts[2] == "" {
			http.Error(w, "Missing account ID", http.StatusBadRequest)
			return
		}
		accountID := parts[2]

		// Mock account data database query
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AccountResponse{
			ID:      accountID,
			Balance: 5742.89, // Mock standard balance
			Status:  "active",
		})
		log.Printf("[Account] Retrieved details for account: %s", accountID)
	})

	log.Printf("[Account] Starting Account Service on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start Account Service: %v", err)
	}
}
