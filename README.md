# Go Scraper API

A high-performance, concurrent web scraping API written in Go. Supports both static HTML scraping and JavaScript-rendered content (React, Vue, Angular SPAs) via headless Chrome.

## Features

- **Ultra-fast concurrent scraping** - 10 parallel workers by default
- **Dual-mode scraping**:
  - Fast HTTP mode for static sites
  - Headless Chrome mode for JavaScript-heavy sites (React, SPAs)
- **Clean text extraction** - Removes scripts, styles, and normalizes whitespace
- **Detailed metrics** - Per-URL timing, success rates, and aggregate statistics
- **Lightweight & efficient** - Minimal memory footprint with optimized HTTP client

## Installation

### Prerequisites

- Go 1.21 or higher
- Google Chrome or Chromium (required only for `render_js` mode)

### Setup

```bash
# Clone the repository
git clone <repository-url>
cd go-scraper-api

# Install dependencies
go mod tidy

# Run the server
go run cmd/scraper-api/main.go
```

The server starts on port **8080**.

## API Usage

### Endpoint

```
POST /scrape
Content-Type: application/json
```

### Request Format

```json
{
    "urls": ["https://example.com", "https://another-site.com"],
    "render_js": false,
    "wait_time_ms": 2000
}
```

**Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `urls` | array | Yes | List of URLs to scrape |
| `render_js` | boolean | No | Use headless Chrome for JS-rendered sites (default: `false`) |
| `wait_time_ms` | integer | No | Wait time for JS rendering in milliseconds (default: `2000`) |

### Response Format

```json
{
    "success": true,
    "summary": {
        "total": 2,
        "successful": 2,
        "failed": 0,
        "total_time_seconds": 0.845
    },
    "results": [
        {
            "url": "https://example.com",
            "success": true,
            "content": "Clean extracted text content...",
            "extraction_time_seconds": 0.423,
            "timestamp": "2026-04-04T21:30:00Z"
        },
        {
            "url": "https://another-site.com",
            "success": true,
            "content": "More extracted content...",
            "extraction_time_seconds": 0.389,
            "timestamp": "2026-04-04T21:30:00Z"
        }
    ]
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `success` | boolean | `true` if all URLs scraped successfully |
| `summary.total` | integer | Total number of URLs requested |
| `summary.successful` | integer | Number of successfully scraped URLs |
| `summary.failed` | integer | Number of failed scrapes |
| `summary.total_time_seconds` | float | Total time for entire operation |
| `results[].url` | string | The scraped URL |
| `results[].success` | boolean | Whether this URL was successfully scraped |
| `results[].content` | string | Extracted clean text content |
| `results[].extraction_time_seconds` | float | Time taken for this URL |
| `results[].timestamp` | string | ISO 8601 timestamp of extraction |
| `results[].error` | string | Error message if `success` is `false` |

## Usage Examples

### 1. Static Website Scraping (Fast Mode)

For static HTML sites like blogs, documentation, news sites:

```bash
curl -X POST http://localhost:8080/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "https://example.com",
      "https://golang.org"
    ]
  }'
```

**Speed:** ~200-500ms per URL

### 2. React/SPA Scraping (Headless Mode)

For JavaScript-heavy sites like React, Vue, Angular applications:

```bash
curl -X POST http://localhost:8080/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "urls": ["https://react-website.com"],
    "render_js": true,
    "wait_time_ms": 3000
  }'
```

**Speed:** ~2-5 seconds per URL (depends on `wait_time_ms`)

### 3. Mixed Batch Scraping

```bash
curl -X POST http://localhost:8080/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "https://static-site.com",
      "https://react-app.com",
      "https://vue-website.com"
    ],
    "render_js": true,
    "wait_time_ms": 2500
  }'
```

## Configuration

### HTTP Client Settings

The scraper uses an optimized HTTP client with connection pooling:

```go
MaxIdleConns:        30
MaxConnsPerHost:     10
MaxIdleConnsPerHost: 10
Timeout:             10 seconds
```

### Headless Chrome Options

When `render_js: true`, Chrome runs with these flags:
- `--headless` - No GUI
- `--disable-gpu` - Disable GPU acceleration
- `--no-sandbox` - Required for Docker/containers
- `--disable-dev-shm-usage` - Prevent /dev/shm issues

## Architecture

```
cmd/scraper-api/
  └── main.go           # Server entry point

internal/
  ├── scraper/
  │   ├── models.go     # Request/response types
  │   ├── scraper.go    # HTTP scraping logic
  │   └── headless.go   # Chrome headless scraping
  └── http/
      └── handlers.go   # HTTP handlers

pkg/
  └── utils/
      └── logger.go     # Logging utilities
```

## Performance

| Mode | Speed | Memory | Use Case |
|------|-------|--------|----------|
| HTTP Mode | ~200-500ms/URL | ~10MB | Static sites |
| Headless Mode | ~2-5s/URL | ~100MB+ | React/SPAs |

## Error Handling

Failed scrapes return partial results with error details:

```json
{
    "success": false,
    "summary": {
        "total": 2,
        "successful": 1,
        "failed": 1,
        "total_time_seconds": 5.234
    },
    "results": [
        {
            "url": "https://example.com",
            "success": true,
            "content": "Extracted content...",
            "extraction_time_seconds": 0.423,
            "timestamp": "2026-04-04T21:30:00Z"
        },
        {
            "url": "https://invalid-domain-12345.com",
            "success": false,
            "content": "",
            "extraction_time_seconds": 10.001,
            "timestamp": "2026-04-04T21:30:10Z",
            "error": "Get \"https://invalid-domain-12345.com\": dial tcp: lookup invalid-domain-12345.com: no such host"
        }
    ]
}
```

## Development

### Building

```bash
go build -o scraper-api cmd/scraper-api/main.go
```

### Testing

```bash
# Run server
go run cmd/scraper-api/main.go

# Test in another terminal
curl -X POST http://localhost:8080/scrape \
  -d '{"urls": ["https://example.com"]}'
```

## License

MIT License
