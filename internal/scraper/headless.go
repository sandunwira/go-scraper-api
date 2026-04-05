package scraper

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// getChromePath finds or downloads Chrome for headless browsing
func getChromePath() (string, error) {
	// Priority 1: Environment variable
	if envPath := os.Getenv("CHROME_PATH"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			fmt.Printf("Using Chrome from CHROME_PATH: %s\n", envPath)
			return envPath, nil
		}
	}

	// Priority 2: Check if chrome exists in current directory (downloaded at runtime)
	localChrome := "./chrome-linux/chrome"
	if _, err := os.Stat(localChrome); err == nil {
		absPath, _ := filepath.Abs(localChrome)
		fmt.Printf("Using local Chrome: %s\n", absPath)
		return absPath, nil
	}

	// Priority 3: Common system paths
	candidates := []string{
		"/opt/render/.chrome/chrome-linux/chrome",
		"/usr/bin/chromium-browser",
		"/usr/bin/chromium",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Using system Chrome: %s\n", path)
			return path, nil
		}
	}

	// Priority 4: Try to download Chrome
	if runtime.GOOS == "linux" {
		fmt.Println("Chrome not found, attempting to download...")
		return downloadChrome()
	}

	return "", fmt.Errorf("Chrome not found and unable to download for OS: %s", runtime.GOOS)
}

// downloadChrome downloads and extracts Chrome to local directory
func downloadChrome() (string, error) {
	chromeDir := "./chrome-linux"
	chromeBin := filepath.Join(chromeDir, "chrome")

	// Check if already exists
	if _, err := os.Stat(chromeBin); err == nil {
		absPath, _ := filepath.Abs(chromeBin)
		return absPath, nil
	}

	// Download Chrome
	fmt.Println("Downloading Chromium...")
	url := "https://storage.googleapis.com/chromium-browser-snapshots/Linux_x64/1097615/chrome-linux.zip"

	cmd := exec.Command("curl", "-sL", url, "-o", "chrome.zip")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to download Chrome: %v", err)
	}

	// Extract
	fmt.Println("Extracting Chromium...")
	cmd = exec.Command("unzip", "-q", "chrome.zip")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to extract Chrome: %v", err)
	}

	// Clean up zip
	os.Remove("chrome.zip")

	// Make executable
	os.Chmod(chromeBin, 0755)

	absPath, _ := filepath.Abs(chromeBin)
	fmt.Printf("Chrome downloaded to: %s\n", absPath)
	return absPath, nil
}

// fetchAndParseWithHeadless uses Chrome headless browser to render JavaScript
func fetchAndParseWithHeadless(url string, waitTimeMs int) ScrapingResult {
	startTime := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get Chrome path (find or download)
	chromePath, err := getChromePath()
	if err != nil {
		return ScrapingResult{
			URL:                   url,
			Success:               false,
			Content:               "",
			ExtractionTimeSeconds: time.Since(startTime).Seconds(),
			Timestamp:             time.Now(),
			Error:                 fmt.Sprintf("Chrome not available: %v", err),
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

	err = chromedp.Run(taskCtx,
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
