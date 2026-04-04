package scraper

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        30,
		MaxConnsPerHost:     10,
		MaxIdleConnsPerHost: 10,
	},
	Timeout: 10 * time.Second,
}

func fetchAndParse(url string) ScrapingResult {
	startTime := time.Now()

	resp, err := httpClient.Get(url)
	if err != nil {
		return ScrapingResult{
			URL:                   url,
			Success:               false,
			Content:               "",
			ExtractionTimeSeconds: time.Since(startTime).Seconds(),
			Timestamp:             time.Now(),
			Error:                 err.Error(),
		}
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ScrapingResult{
			URL:                   url,
			Success:               false,
			Content:               "",
			ExtractionTimeSeconds: time.Since(startTime).Seconds(),
			Timestamp:             time.Now(),
			Error:                 err.Error(),
		}
	}

	// Remove script and style elements
	doc.Find("script, style, noscript, iframe, canvas, svg").Remove()

	// Get clean text from body
	text := doc.Find("body").Text()

	// Clean up whitespace
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")

	extractionTime := time.Since(startTime).Seconds()

	return ScrapingResult{
		URL:                   url,
		Success:               true,
		Content:               text,
		ExtractionTimeSeconds: extractionTime,
		Timestamp:             time.Now(),
	}
}

func ScrapeMany(req ScrapingRequest) ScrapingResponse {
	const numWorkers = 10
	startTime := time.Now()
	jobs := make(chan string, len(req.URLs))
	results := make(chan ScrapingResult, len(req.URLs))

	var wg sync.WaitGroup

	// Choose scraper function based on RenderJS option
	scrapeFunc := fetchAndParse
	if req.RenderJS {
		scrapeFunc = func(url string) ScrapingResult {
			return fetchAndParseWithHeadless(url, req.WaitTime)
		}
	}

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range jobs {
				results <- scrapeFunc(url)
			}
		}()
	}

	// Send jobs
	go func() {
		for _, url := range req.URLs {
			jobs <- url
		}
		close(jobs)
	}()

	// Close results when all done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var resultSlice []ScrapingResult
	successful := 0
	failed := 0

	for r := range results {
		resultSlice = append(resultSlice, r)
		if r.Success {
			successful++
		} else {
			failed++
		}
	}

	totalTime := time.Since(startTime).Seconds()

	allSuccess := failed == 0 && len(req.URLs) > 0

	return ScrapingResponse{
		Success: allSuccess,
		Summary: Summary{
			Total:            len(req.URLs),
			Successful:       successful,
			Failed:           failed,
			TotalTimeSeconds: totalTime,
		},
		Results: resultSlice,
	}
}
