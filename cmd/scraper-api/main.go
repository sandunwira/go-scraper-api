package main

import (
	"log"
	"net/http"

	httpHandlers "go-scraper-api/internal/http"
)

func main() {
	http.HandleFunc("/scrape", httpHandlers.ScrapeHandler)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
