# API Reference

Complete API documentation for the GoURL URL Shortener service.

## Base URL

```
http://localhost:8080
```

For production, replace with your actual domain.

---

## Endpoints

### 1. Health Check

Check if the service is running and healthy.

**Request:**
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy"
}
```

**Status Codes:**
- `200 OK` - Service is healthy

---

### 2. Create Short URL

Create a new short URL from a long URL.

**Request:**
```http
POST /api/shorten
Content-Type: application/json

{
  "url": "https://www.example.com/very/long/path/to/resource",
  "custom_code": "optional-code"  // Optional
}
```

**Request Fields:**
- `url` (string, required): The long URL to shorten. Must be a valid HTTP/HTTPS URL.
- `custom_code` (string, optional): Custom short code (3-20 alphanumeric characters). If not provided, a random 6-character code is generated.

**Success Response:**
```json
{
  "short_code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "long_url": "https://www.example.com/very/long/path/to/resource"
}
```

**Response Fields:**
- `short_code` (string): The generated or custom short code
- `short_url` (string): The complete short URL
- `long_url` (string): The original long URL

**Status Codes:**
- `201 Created` - Short URL created successfully
- `400 Bad Request` - Invalid input
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

**Error Response:**
```json
{
  "error": "error message"
}
```

**Error Cases:**

| Error Message | Cause | Status |
|--------------|-------|--------|
| `url is required` | Missing `url` field | 400 |
| `invalid URL format` | URL is not a valid HTTP/HTTPS URL | 400 |
| `invalid custom code: must be alphanumeric and 3-20 characters` | Custom code doesn't meet requirements | 400 |
| `custom code already exists` | Custom code is already in use | 400 |
| `rate limit exceeded` | Too many requests | 429 |
| `internal server error` | Server-side error | 500 |

**Examples:**

```bash
# Create with auto-generated code
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.google.com"}'

# Create with custom code
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com", "custom_code": "mypage"}'

# Invalid URL
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "not-a-url"}'
# Returns: {"error":"invalid URL format"}
```

---

### 3. Redirect to Original URL

Access a short URL and get redirected to the original URL.

**Request:**
```http
GET /{short_code}
```

**Parameters:**
- `short_code` (path parameter): The short code to redirect

**Response:**
- HTTP 301 redirect with `Location` header set to the original URL
- Click count is incremented asynchronously

**Status Codes:**
- `301 Moved Permanently` - Successful redirect
- `404 Not Found` - Short code doesn't exist
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

**Error Response:**
```json
{
  "error": "error message"
}
```

**Error Cases:**

| Error Message | Cause | Status |
|--------------|-------|--------|
| `short code is required` | Empty short code | 400 |
| `URL not found` | Short code doesn't exist | 404 |
| `rate limit exceeded` | Too many requests | 429 |
| `internal server error` | Server-side error | 500 |

**Examples:**

```bash
# Follow redirect (curl -L follows redirects)
curl -L http://localhost:8080/abc123

# See redirect without following
curl -I http://localhost:8080/abc123
# Returns: HTTP/1.1 301 Moved Permanently
#          Location: https://www.example.com

# Non-existent code
curl http://localhost:8080/nonexistent
# Returns: {"error":"URL not found"}
```

---

## Rate Limiting

All endpoints are rate-limited per client IP address.

**Default Limits:**
- 10 requests per 60 seconds per IP

**Rate Limit Response:**
```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json

{
  "error": "rate limit exceeded"
}
```

**Rate Limit Behavior:**
- Counter starts on first request
- Resets after the time window expires
- Tracked separately per IP address
- Uses sliding window algorithm

**Configuration:**
Set via environment variables:
- `RATE_LIMIT_REQUESTS`: Number of requests allowed
- `RATE_LIMIT_WINDOW`: Time window in seconds

---

## Common Response Headers

All responses include:
```
Content-Type: application/json
```

Redirect responses include:
```
Location: <original_url>
```

---

## Error Handling

All errors follow this format:
```json
{
  "error": "human-readable error message"
}
```

**HTTP Status Code Reference:**

| Status | Meaning | When Used |
|--------|---------|-----------|
| 200 | OK | Successful GET requests (health check) |
| 201 | Created | URL successfully shortened |
| 301 | Moved Permanently | Successful redirect |
| 400 | Bad Request | Invalid input or request |
| 404 | Not Found | Short code doesn't exist |
| 405 | Method Not Allowed | Wrong HTTP method used |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server-side error |

---

## Client IP Detection

Client IP is detected in this order:
1. `X-Forwarded-For` header (first IP)
2. `X-Real-IP` header
3. Direct connection IP

When behind a proxy/load balancer, ensure proper headers are set.

---

## URL Validation

URLs must meet these requirements:
- Must include a scheme (http or https)
- Must include a host
- Only http and https schemes are allowed
- Examples:
  - ✅ `https://www.example.com`
  - ✅ `http://example.com/path?query=value`
  - ❌ `www.example.com` (missing scheme)
  - ❌ `ftp://example.com` (invalid scheme)
  - ❌ `not-a-url` (invalid format)

---

## Short Code Format

**Auto-generated codes:**
- Length: 6 characters (configurable)
- Characters: a-z, A-Z, 0-9 (62 possible characters)
- Cryptographically secure random generation
- Example: `r1ZBjJ`, `KOElLE`, `3sJMxu`

**Custom codes:**
- Length: 3-20 characters
- Characters: a-z, A-Z, 0-9 only
- Must be unique
- Case-sensitive
- Examples: `mypage`, `docs2024`, `GitHub`

---

## Caching Behavior

**Cache Strategy:**
- Cache-aside (lazy loading)
- URLs are cached on:
  - First creation
  - First access after cache miss
- Cache TTL: 1 hour (default, configurable)

**Cache Keys:**
- Format: `{short_code}`
- Example: `abc123`

**Cache Invalidation:**
- Automatic via TTL
- Can be manually cleared via Redis CLI if needed

---

## Example Workflows

### Basic URL Shortening Workflow

```bash
# 1. Create short URL
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.verylongdomainname.com/very/long/path/to/resource"}'

# Response:
# {
#   "short_code": "r1ZBjJ",
#   "short_url": "http://localhost:8080/r1ZBjJ",
#   "long_url": "https://www.verylongdomainname.com/very/long/path/to/resource"
# }

# 2. Use the short URL
curl -L http://localhost:8080/r1ZBjJ
# Redirects to: https://www.verylongdomainname.com/very/long/path/to/resource
```

### Custom Short Code Workflow

```bash
# 1. Create with custom code
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://docs.example.com", "custom_code": "docs"}'

# Response:
# {
#   "short_code": "docs",
#   "short_url": "http://localhost:8080/docs",
#   "long_url": "https://docs.example.com"
# }

# 2. Access via custom code
curl -L http://localhost:8080/docs
# Redirects to: https://docs.example.com
```

### Error Handling Workflow

```bash
# 1. Try to create duplicate custom code
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "custom_code": "docs"}'

# Response (400):
# {"error":"custom code already exists"}

# 2. Try invalid URL
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "not-a-url"}'

# Response (400):
# {"error":"invalid URL format"}

# 3. Access non-existent code
curl http://localhost:8080/notfound

# Response (404):
# {"error":"URL not found"}
```

---

## Testing with Different HTTP Clients

### curl
```bash
# Create URL
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'

# Get redirect (follow)
curl -L http://localhost:8080/abc123

# Get redirect (don't follow)
curl -v http://localhost:8080/abc123
```

### HTTPie
```bash
# Create URL
http POST http://localhost:8080/api/shorten url=https://example.com

# Get redirect
http --follow http://localhost:8080/abc123
```

### JavaScript (fetch)
```javascript
// Create URL
fetch('http://localhost:8080/api/shorten', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({url: 'https://example.com'})
})
.then(res => res.json())
.then(data => console.log(data));

// Access short URL (will follow redirect)
fetch('http://localhost:8080/abc123')
.then(res => console.log(res.url)); // Final URL after redirect
```

### Python (requests)
```python
import requests

# Create URL
response = requests.post(
    'http://localhost:8080/api/shorten',
    json={'url': 'https://example.com'}
)
print(response.json())

# Access short URL
response = requests.get(
    'http://localhost:8080/abc123',
    allow_redirects=True
)
print(response.url)  # Final URL after redirect
```

---

## Performance Considerations

**Response Times (typical):**
- Health check: <5ms
- URL creation: 10-50ms (cache + DB write)
- URL redirect (cache hit): <10ms
- URL redirect (cache miss): 20-50ms (DB read + cache write)

**Throughput:**
- Depends on hardware and configuration
- With default settings:
  - Limited by rate limiting: 10 req/min per IP
  - Without rate limits: hundreds of requests per second

**Scalability:**
- Stateless application (horizontal scaling)
- Redis for distributed caching
- PostgreSQL with connection pooling
- Can add read replicas for reads

---

## Security Considerations

1. **Input Validation**: All inputs are validated
2. **SQL Injection**: Prevented via parameterized queries
3. **Rate Limiting**: Prevents abuse
4. **HTTPS Only**: Only http/https URLs accepted
5. **No XSS**: JSON API (not rendering HTML)

---

## Monitoring

**Health Check:**
```bash
# Simple health check
curl http://localhost:8080/health
```

**Database Check:**
```bash
# Check database connection
docker exec gourl_postgres pg_isready -U urlshortener
```

**Redis Check:**
```bash
# Check Redis connection
docker exec gourl_redis redis-cli ping
```

---

## Troubleshooting

**Problem: Rate limited**
- Wait for the rate limit window to expire (default: 60 seconds)
- Or adjust `RATE_LIMIT_REQUESTS` and `RATE_LIMIT_WINDOW` environment variables

**Problem: Can't connect**
- Check if service is running: `curl http://localhost:8080/health`
- Check Docker containers: `docker ps`
- Check logs: `docker logs gourl_app`

**Problem: Database errors**
- Check PostgreSQL is running: `docker ps | grep postgres`
- Check database logs: `docker logs gourl_postgres`
- Verify connection settings in `.env`

**Problem: Cache not working**
- Check Redis is running: `docker ps | grep redis`
- Test Redis: `docker exec gourl_redis redis-cli ping`
- Check Redis logs: `docker logs gourl_redis`

---

## Version

API Version: 1.0.0  
Last Updated: December 2025
