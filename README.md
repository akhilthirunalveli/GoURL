# GoURL - URL Shortener Service

A high-performance URL shortener backend built with Go, PostgreSQL, and Redis. This service provides REST APIs to create short links and redirect users to original URLs with low latency.

## Features

- 🚀 **Fast URL shortening** with unique short code generation
- 🔄 **Low-latency redirects** with Redis caching
- 🛡️ **Rate limiting** to prevent abuse
- 💾 **PostgreSQL** for persistent storage
- ⚡ **Redis** for caching and rate limiting
- 🔒 **Safe concurrent handling** with proper synchronization
- ⚙️ **Environment-based configuration**
- 🐳 **Docker support** for local development
- ✅ **Proper HTTP status codes** for all responses
- 📊 **Click tracking** for analytics

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Go API    │
│  (Handlers) │
└──────┬──────┘
       │
       ├──────────────────┐
       ▼                  ▼
┌─────────────┐    ┌─────────────┐
│    Redis    │    │ PostgreSQL  │
│  (Cache &   │    │  (Persist)  │
│Rate Limit)  │    │             │
└─────────────┘    └─────────────┘
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development without Docker)

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/akhilthirunalveli/GoURL.git
cd GoURL
```

2. Start the services:
```bash
docker-compose up --build
```

3. The API will be available at `http://localhost:8080`

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL and Redis (or use Docker Compose for just these services):
```bash
docker-compose up postgres redis
```

3. Copy the example environment file:
```bash
cp .env.example .env
```

4. Update `.env` with your configuration (use `localhost` for local development)

5. Run the application:
```bash
go run cmd/server/main.go
```

## API Endpoints

### 1. Create Short URL

**Endpoint:** `POST /api/shorten`

**Request Body:**
```json
{
  "url": "https://www.example.com/very/long/url",
  "custom_code": "optional-custom-code"
}
```

**Success Response (201 Created):**
```json
{
  "short_code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "long_url": "https://www.example.com/very/long/url"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid URL or custom code already exists
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

### 2. Redirect to Original URL

**Endpoint:** `GET /{short_code}`

**Success Response:**
- `301 Moved Permanently` - Redirects to the original URL

**Error Responses:**
- `404 Not Found` - Short code doesn't exist
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

### 3. Health Check

**Endpoint:** `GET /health`

**Success Response (200 OK):**
```json
{
  "status": "healthy"
}
```

## Configuration

Configuration is done through environment variables. See `.env.example` for all available options:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Server port | 8080 |
| `SERVER_HOST` | Server host | 0.0.0.0 |
| `DB_HOST` | PostgreSQL host | localhost |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_USER` | Database user | urlshortener |
| `DB_PASSWORD` | Database password | urlshortener_pass |
| `DB_NAME` | Database name | urlshortener |
| `REDIS_HOST` | Redis host | localhost |
| `REDIS_PORT` | Redis port | 6379 |
| `BASE_URL` | Base URL for short links | http://localhost:8080 |
| `SHORT_CODE_LENGTH` | Length of generated codes | 6 |
| `RATE_LIMIT_REQUESTS` | Requests per window | 10 |
| `RATE_LIMIT_WINDOW` | Rate limit window (seconds) | 60 |
| `CACHE_TTL` | Cache TTL (seconds) | 3600 |

## Project Structure

```
.
├── cmd/
│   └── server/          # Application entry point
│       └── main.go
├── internal/
│   ├── cache/           # Redis cache implementation
│   │   └── redis.go
│   ├── config/          # Configuration management
│   │   └── config.go
│   ├── database/        # PostgreSQL database layer
│   │   └── postgres.go
│   ├── handler/         # HTTP handlers
│   │   └── url_handler.go
│   ├── models/          # Data models
│   │   └── url.go
│   └── service/         # Business logic
│       └── url_service.go
├── docker-compose.yml   # Docker Compose configuration
├── Dockerfile           # Application Dockerfile
├── init.sql            # Database schema
├── .env.example        # Example environment variables
└── README.md           # This file
```

## Key Features Explained

### 1. Unique Short Code Generation

The service generates cryptographically secure random short codes using Go's `crypto/rand` package. If a collision occurs (rare), it retries up to 10 times. Custom codes are also supported with validation.

### 2. Caching Strategy

- **Cache-aside pattern**: Check cache first, then database
- **Write-through**: URLs are cached when created or first accessed
- **TTL-based expiration**: Configurable cache expiration
- **Graceful degradation**: Service continues if cache is unavailable

### 3. Rate Limiting

- **Per-IP rate limiting** using Redis counters
- **Sliding window algorithm** for fair rate limiting
- **Configurable limits** via environment variables
- **Proper HTTP 429 responses** when limit exceeded

### 4. Concurrency Safety

- **Mutex protection** for short code generation
- **Database connection pooling** with proper limits
- **Goroutines** for async click count updates
- **Context-based timeouts** for external calls

### 5. Error Handling

All endpoints return appropriate HTTP status codes:
- `200 OK` - Successful GET requests
- `201 Created` - URL successfully created
- `301 Moved Permanently` - Successful redirect
- `400 Bad Request` - Invalid input
- `404 Not Found` - Resource not found
- `405 Method Not Allowed` - Wrong HTTP method
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server errors

## Testing

### Manual Testing with cURL

Create a short URL:
```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com"}'
```

Access the short URL:
```bash
curl -L http://localhost:8080/abc123
```

Test with custom code:
```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com", "custom_code": "mycode"}'
```

Health check:
```bash
curl http://localhost:8080/health
```

### Testing Rate Limiting

Run multiple requests quickly to trigger rate limiting:
```bash
for i in {1..15}; do
  curl -X POST http://localhost:8080/api/shorten \
    -H "Content-Type: application/json" \
    -d "{\"url\": \"https://www.example.com/page$i\"}"
  echo ""
done
```

## Building

### Build Binary

```bash
go build -o bin/server cmd/server/main.go
```

### Build Docker Image

```bash
docker build -t gourl:latest .
```

## Production Considerations

1. **Use HTTPS** in production (update `BASE_URL`)
2. **Set strong database passwords**
3. **Configure proper Redis persistence**
4. **Set up monitoring** for rate limits and errors
5. **Use environment-specific configs**
6. **Implement logging** with structured logs
7. **Add metrics** for observability
8. **Set up database backups**
9. **Use a reverse proxy** (nginx, etc.) for TLS termination
10. **Implement proper CORS** if needed

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the MIT License.

## Acknowledgments

Built with:
- [Go](https://golang.org/) - Programming language
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Redis](https://redis.io/) - Cache and rate limiting
- [Docker](https://www.docker.com/) - Containerization 
