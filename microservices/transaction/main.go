package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type TransactionRequest struct {
	SenderID   string  `json:"sender_id"`
	ReceiverID string  `json:"receiver_id"`
	Amount     float64 `json:"amount"`
}

type TransactionResponse struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

var notificationServiceURL string

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	notificationServiceURL = os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationServiceURL == "" {
		notificationServiceURL = "http://localhost:8084"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthResponse{Status: "UP"})
	})

	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		txnID := fmt.Sprintf("txn-%d", time.Now().UnixNano())
		resp := TransactionResponse{
			TransactionID: txnID,
			Status:        "completed",
			Timestamp:     time.Now(),
		}

		// Asynchronously trigger notification
		go func(tID string, amount float64) {
			payload := map[string]interface{}{
				"event":   "transaction_completed",
				"message": fmt.Sprintf("Transaction %s of $%.2f was successful", tID, amount),
			}
			data, err := json.Marshal(payload)
			if err != nil {
				log.Printf("[Transaction] Failed to marshal notification payload: %v", err)
				return
			}

			client := http.Client{Timeout: 3 * time.Second}
			resp, err := client.Post(notificationServiceURL+"/notify", "application/json", bytes.NewBuffer(data))
			if err != nil {
				log.Printf("[Transaction] Failed to dispatch notification: %v", err)
				return
			}
			defer resp.Body.Close()
			log.Printf("[Transaction] Dispatched notification for %s, response code: %d", tID, resp.StatusCode)
		}(txnID, req.Amount)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		log.Printf("[Transaction] Created transaction: %s", txnID)
	})

	log.Printf("[Transaction] Starting Transaction Service on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start Transaction Service: %v", err)
	}
}
