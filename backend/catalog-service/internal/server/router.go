package server

import (
	"encoding/json"
	"net/http"
	"time"

	"video-streaming/backend/catalog-service/internal/handler"
	"video-streaming/backend/catalog-service/internal/repository"
)

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

// NewRouter builds all catalog-service routes.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	catalogHandler := handler.NewHTTPHandler(repository.NewInMemoryStore())
	mux.Handle("/internal/v1/catalog/tracks/", catalogHandler)
	mux.HandleFunc("/healthz", healthzHandler)
	return mux
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := healthResponse{
		Status:    "ok",
		Service:   "catalog-service",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
