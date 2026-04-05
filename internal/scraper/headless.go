package scraper

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// fetchAndParseWithHeadless uses Chrome headless browser to render JavaScript
func fetchAndParseWithHeadless(url string, waitTimeMs int) ScrapingResult {
	startTime := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create Chrome instance with custom path if available
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
	)

	// Check for custom Chrome/Chromium path from environment
	if chromePath := os.Getenv("CHROME_PATH"); chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var html string

	// Default wait time
	if waitTimeMs == 0 {
		waitTimeMs = 2000 // 2 seconds default
	}

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.Sleep(time.Duration(waitTimeMs)*time.Millisecond),
		chromedp.OuterHTML("html", &html),
	)

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

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
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
