package scraper

import "time"

type ScrapingRequest struct {
	URLs     []string `json:"urls"`
	RenderJS bool     `json:"render_js,omitempty"`
	WaitTime int      `json:"wait_time_ms,omitempty"`
}

type ScrapingResult struct {
	URL                   string    `json:"url"`
	Success               bool      `json:"success"`
	Content               string    `json:"content,omitempty"`
	ExtractionTimeSeconds float64   `json:"extraction_time_seconds"`
	Timestamp             time.Time `json:"timestamp"`
	Error                 string    `json:"error,omitempty"`
}

type Summary struct {
	Total            int     `json:"total"`
	Successful       int     `json:"successful"`
	Failed           int     `json:"failed"`
	TotalTimeSeconds float64 `json:"total_time_seconds"`
}

type ScrapingResponse struct {
	Success bool             `json:"success"`
	Summary Summary          `json:"summary"`
	Results []ScrapingResult `json:"results"`
}
