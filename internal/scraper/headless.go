package scraper

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// findChromePath looks for Chrome/Chromium in common locations
func findChromePath() string {
	// Check environment variable first
	if path := os.Getenv("CHROME_PATH"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Common Chrome/Chromium paths
	candidates := []string{
		"/opt/render/.chrome/chrome-linux/chrome",
		"/usr/bin/chromium-browser",
		"/usr/bin/chromium",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// fetchAndParseWithHeadless uses Chrome headless browser to render JavaScript
func fetchAndParseWithHeadless(url string, waitTimeMs int) ScrapingResult {
	startTime := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find Chrome path
	chromePath := findChromePath()
	if chromePath == "" {
		return ScrapingResult{
			URL:                   url,
			Success:               false,
			Content:               "",
			ExtractionTimeSeconds: time.Since(startTime).Seconds(),
			Timestamp:             time.Now(),
			Error:                 "Chrome/Chromium not found. Set CHROME_PATH environment variable.",
		}
	}

	// Create Chrome instance with custom path
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("single-process", true),
		chromedp.Flag("no-zygote", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
	)

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
