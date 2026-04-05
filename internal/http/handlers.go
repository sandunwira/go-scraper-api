package http

import (
	"encoding/json"
	"net/http"

	"go-scraper-api/internal/scraper"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"message": "Hello from Crawler!",
	}
	json.NewEncoder(w).Encode(response)
}

func ScrapeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req scraper.ScrapingRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.URLs) == 0 {
		http.Error(w, "empty urls list", http.StatusBadRequest)
		return
	}

	// Concurrent scrape of all URLs in this request
	response := scraper.ScrapeMany(req)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "response encoding failed", http.StatusInternalServerError)
	}
}
