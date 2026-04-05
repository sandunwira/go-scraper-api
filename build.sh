#!/bin/bash

# Build script for Render.com to install Chromium (without apt-get)

set -e

CHROME_DIR="/opt/render/.chrome"
CHROME_BIN="$CHROME_DIR/chrome-linux/chrome"

# Check if Chrome already exists (cached from previous build)
if [ -f "$CHROME_BIN" ]; then
    echo "Chromium already exists at $CHROME_BIN"
    $CHROME_BIN --version || true
else
    echo "Downloading Chromium for headless scraping..."
    
    # Create directory for Chrome
    mkdir -p "$CHROME_DIR"
    cd "$CHROME_DIR"
    
    # Download Chromium (stable version for Linux x64)
    CHROME_URL="https://storage.googleapis.com/chromium-browser-snapshots/Linux_x64/1097615/chrome-linux.zip"
    
    echo "Downloading Chromium from $CHROME_URL..."
    curl -sL "$CHROME_URL" -o chrome-linux.zip
    
    echo "Extracting Chromium..."
    unzip -q chrome-linux.zip
    
    echo "Setting permissions..."
    chmod +x chrome-linux/chrome
    
    # Verify
    echo "Chromium installed at: $CHROME_BIN"
    $CHROME_BIN --version || true
    
    # Go back to project root
    cd "$OLDPWD"
fi

# Export path for the build
export CHROME_PATH="$CHROME_BIN"
echo "CHROME_PATH=$CHROME_PATH"

# Build the Go application
echo "Building Go application..."
go build -o app cmd/scraper-api/main.go

echo "Build complete!"
