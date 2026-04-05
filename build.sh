#!/bin/bash

# Build script for Render.com to install Chromium

echo "Installing Chromium for headless scraping..."

# Update package lists
apt-get update -qq

# Install Chromium (headless browser)
apt-get install -y -qq chromium-browser

# Verify installation
if command -v chromium-browser &> /dev/null; then
    echo "Chromium installed successfully at: $(which chromium-browser)"
    chromium-browser --version
else
    echo "Failed to install Chromium"
    exit 1
fi

# Set environment variable for chromedp to find Chrome
export CHROME_PATH=$(which chromium-browser)
echo "CHROME_PATH=$CHROME_PATH"

# Build the Go application
echo "Building Go application..."
go build -o scraper-api cmd/scraper-api/main.go

echo "Build complete!"
