package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ValidateResponse struct {
	Valid bool `json:"valid"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthResponse{Status: "UP"})
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Simple mock auth check
		if req.Username == "admin" && req.Password == "password" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(LoginResponse{Token: "mock-jwt-token-123"})
			log.Printf("[Auth] Successfully authenticated user: %s", req.Username)
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Printf("[Auth] Failed authentication attempt for user: %s", req.Username)
		}
	})

	http.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		token := r.Header.Get("Authorization")
		if token == "" {
			// Also fallback to body for flexibility
			var bodyReq struct {
				Token string `json:"token"`
			}
			if err := json.NewDecoder(r.Body).Decode(&bodyReq); err == nil {
				token = bodyReq.Token
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if token == "Bearer mock-jwt-token-123" || token == "mock-jwt-token-123" {
			json.NewEncoder(w).Encode(ValidateResponse{Valid: true})
			log.Println("[Auth] Token validated successfully")
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ValidateResponse{Valid: false})
			log.Println("[Auth] Token validation failed")
		}
	})

	log.Printf("[Auth] Starting Auth Service on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start Auth Service: %v", err)
	}
}
