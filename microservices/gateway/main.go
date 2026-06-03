package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

type HealthStatus struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

type GatewayHealthResponse struct {
	Status   string         `json:"status"`
	Services []HealthStatus `json:"services"`
}

type AuthValidateResponse struct {
	Valid bool `json:"valid"`
}

var (
	authServiceURL   string
	accountServiceURL string
	txnServiceURL     string
	notifServiceURL   string
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	authServiceURL = os.Getenv("AUTH_SERVICE_URL")
	accountServiceURL = os.Getenv("ACCOUNT_SERVICE_URL")
	txnServiceURL = os.Getenv("TRANSACTION_SERVICE_URL")
	notifServiceURL = os.Getenv("NOTIFICATION_SERVICE_URL")

	// Set defaults if empty (for local testing)
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8081"
	}
	if accountServiceURL == "" {
		accountServiceURL = "http://localhost:8082"
	}
	if txnServiceURL == "" {
		txnServiceURL = "http://localhost:8083"
	}
	if notifServiceURL == "" {
		notifServiceURL = "http://localhost:8084"
	}

	// Create reverse proxies
	authProxy := createReverseProxy(authServiceURL, "/login")
	accountProxy := createReverseProxy(accountServiceURL, "")
	txnProxy := createReverseProxy(txnServiceURL, "/transactions")

	// Router handler
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Gateway] Routing login request to Auth Service")
		authProxy.ServeHTTP(w, r)
	})
	http.HandleFunc("/api/v1/accounts/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Gateway] Routing account request - authenticating...")
		if !authenticate(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized - Invalid token"})
			return
		}
		// Rewrite path: /api/v1/accounts/123 -> /accounts/123
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/accounts/")
		r.URL.Path = "/accounts/" + id
		accountProxy.ServeHTTP(w, r)
	})
	http.HandleFunc("/api/v1/transactions", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Gateway] Routing transaction request - authenticating...")
		if !authenticate(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized - Invalid token"})
			return
		}
		txnProxy.ServeHTTP(w, r)
	})

	log.Printf("[Gateway] Starting API Gateway on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Gateway server failed: %v", err)
	}
}

func createReverseProxy(target string, pathOverride string) *httputil.ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid service URL: %s", target)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		if pathOverride != "" {
			req.URL.Path = pathOverride
		}
		req.Host = targetURL.Host
	}
	return proxy
}

func authenticate(r *http.Request) bool {
	token := r.Header.Get("Authorization")
	if token == "" {
		return false
	}

	client := http.Client{Timeout: 2 * time.Second}
	validateReq, err := http.NewRequest(http.MethodPost, authServiceURL+"/validate", nil)
	if err != nil {
		return false
	}
	validateReq.Header.Set("Authorization", token)

	resp, err := client.Do(validateReq)
	if err != nil {
		log.Printf("[Gateway] Auth Validation error: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var authResp AuthValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return false
	}
	return authResp.Valid
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	services := []struct {
		name string
		url  string
	}{
		{"auth-service", authServiceURL},
		{"account-service", accountServiceURL},
		{"transaction-service", txnServiceURL},
		{"notification-service", notifServiceURL},
	}

	client := http.Client{Timeout: 1 * time.Second}
	var serviceStatuses []HealthStatus
	allHealthy := true

	for _, s := range services {
		status := HealthStatus{Service: s.name, Status: "UP"}
		resp, err := client.Get(s.url + "/health")
		if err != nil {
			status.Status = "DOWN"
			status.Error = err.Error()
			allHealthy = false
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				status.Status = "DOWN"
				status.Error = "Non-200 response"
				allHealthy = false
			}
		}
		serviceStatuses = append(serviceStatuses, status)
	}

	w.Header().Set("Content-Type", "application/json")
	if allHealthy {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GatewayHealthResponse{
			Status:   "UP",
			Services: serviceStatuses,
		})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(GatewayHealthResponse{
			Status:   "DOWN",
			Services: serviceStatuses,
		})
	}
}
