# Testing Guide

This document provides comprehensive testing instructions for the GoURL shortener service.

## Prerequisites

Before testing, ensure you have:
- Docker and Docker Compose installed
- Go 1.21+ (for local development)
- curl or similar HTTP client

## Quick Start with Docker Compose

The easiest way to test the application is using Docker Compose:

```bash
# Start all services
docker compose up --build

# In another terminal, run tests
./test.sh
```

## Manual Testing

### 1. Start the Services

Using Docker Compose:
```bash
docker compose up -d
```

Or manually with Go (requires PostgreSQL and Redis running):
```bash
# Start PostgreSQL
docker run -d --name gourl_postgres \
  -e POSTGRES_USER=urlshortener \
  -e POSTGRES_PASSWORD=urlshortener_pass \
  -e POSTGRES_DB=urlshortener \
  -p 5432:5432 postgres:15-alpine

# Initialize database
sleep 5
docker exec gourl_postgres psql -U urlshortener -d urlshortener -c "
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(20) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    click_count INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_created_at ON urls(created_at);
"

# Start Redis
docker run -d --name gourl_redis -p 6379:6379 redis:7-alpine

# Copy and configure environment
cp .env.example .env
# Edit .env if needed

# Run the application
go run cmd/server/main.go
```

### 2. Test Health Check

```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{"status":"healthy"}
```

### 3. Test URL Shortening

#### Create a Short URL
```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.google.com"}'
```

**Expected Response (201 Created):**
```json
{
  "short_code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "long_url": "https://www.google.com"
}
```

#### Create with Custom Code
```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com", "custom_code": "mycode"}'
```

**Expected Response (201 Created):**
```json
{
  "short_code": "mycode",
  "short_url": "http://localhost:8080/mycode",
  "long_url": "https://www.example.com"
}
```

### 4. Test URL Redirection

```bash
curl -v http://localhost:8080/mycode
```

**Expected Response:**
- HTTP Status: `301 Moved Permanently`
- Location header: `https://www.example.com`

### 5. Test Error Handling

#### Invalid URL
```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "not-a-valid-url"}'
```

**Expected Response (400 Bad Request):**
```json
{"error":"invalid URL format"}
```

#### Duplicate Custom Code
```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com", "custom_code": "mycode"}'
```

**Expected Response (400 Bad Request):**
```json
{"error":"custom code already exists"}
```

#### Non-existent Short Code
```bash
curl http://localhost:8080/nonexistent
```

**Expected Response (404 Not Found):**
```json
{"error":"URL not found"}
```

#### Missing URL Field
```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Expected Response (400 Bad Request):**
```json
{"error":"url is required"}
```

### 6. Test Rate Limiting

Send more than 10 requests within 60 seconds:

```bash
for i in {1..12}; do
  echo "Request $i:"
  curl -w " (HTTP: %{http_code})\n" -X POST http://localhost:8080/api/shorten \
    -H "Content-Type: application/json" \
    -d "{\"url\": \"https://www.example.com/page$i\"}"
done
```

**Expected Behavior:**
- First 10 requests: `201 Created`
- Requests 11+: `429 Too Many Requests` with error message:
  ```json
  {"error":"rate limit exceeded"}
  ```

### 7. Test Caching

#### Verify Redis Cache
```bash
# Check that URLs are cached
docker exec gourl_redis redis-cli keys "*"

# Get a specific cached URL
docker exec gourl_redis redis-cli get "mycode"
```

#### Verify Cache Hit Performance
```bash
# First request (cache miss)
time curl -s http://localhost:8080/mycode > /dev/null

# Second request (cache hit - should be faster)
time curl -s http://localhost:8080/mycode > /dev/null
```

### 8. Test Click Tracking

```bash
# Create a URL
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com", "custom_code": "trackme"}'

# Access it multiple times
curl -s http://localhost:8080/trackme > /dev/null
curl -s http://localhost:8080/trackme > /dev/null
curl -s http://localhost:8080/trackme > /dev/null

# Wait a moment for async updates
sleep 2

# Check click count in database
docker exec gourl_postgres psql -U urlshortener -d urlshortener \
  -c "SELECT short_code, click_count FROM urls WHERE short_code = 'trackme';"
```

**Expected Output:**
```
 short_code | click_count 
------------+-------------
 trackme    |           3
```

### 9. Test Concurrent Requests

Test that the service handles concurrent requests safely:

```bash
# Run 20 concurrent requests
seq 1 20 | xargs -P 20 -I {} curl -s -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com/concurrent{}"}'

# Verify all codes are unique
docker exec gourl_postgres psql -U urlshortener -d urlshortener \
  -c "SELECT COUNT(*), COUNT(DISTINCT short_code) FROM urls;"
```

**Expected:** Both counts should be equal (no duplicate short codes)

### 10. Test Different HTTP Methods

#### Wrong Method on Shorten Endpoint
```bash
curl -X GET http://localhost:8080/api/shorten
```

**Expected Response (405 Method Not Allowed):**
```json
{"error":"method not allowed"}
```

#### Wrong Method on Redirect Endpoint
```bash
curl -X POST http://localhost:8080/mycode
```

**Expected Response (405 Method Not Allowed):**
```json
{"error":"method not allowed"}
```

## Database Verification

Check the database state:

```bash
# View all URLs
docker exec gourl_postgres psql -U urlshortener -d urlshortener \
  -c "SELECT short_code, original_url, click_count, created_at FROM urls ORDER BY created_at DESC LIMIT 10;"

# Check for duplicate short codes (should return 0)
docker exec gourl_postgres psql -U urlshortener -d urlshortener \
  -c "SELECT short_code, COUNT(*) FROM urls GROUP BY short_code HAVING COUNT(*) > 1;"

# Check total URLs created
docker exec gourl_postgres psql -U urlshortener -d urlshortener \
  -c "SELECT COUNT(*) as total_urls FROM urls;"
```

## Redis Verification

Check Redis state:

```bash
# List all keys
docker exec gourl_redis redis-cli keys "*"

# Check rate limit for a specific IP
docker exec gourl_redis redis-cli get "ratelimit:127.0.0.1"

# Get TTL of a cached URL
docker exec gourl_redis redis-cli ttl "mycode"

# Check Redis info
docker exec gourl_redis redis-cli info stats
```

## Performance Testing

### Basic Load Test with Apache Bench (if available)

```bash
# Install apache bench if needed
# Ubuntu/Debian: sudo apt-get install apache2-utils
# macOS: brew install httpd (includes ab)

# Test create endpoint (100 requests, 10 concurrent)
ab -n 100 -c 10 -p payload.json -T application/json \
  http://localhost:8080/api/shorten

# Where payload.json contains:
echo '{"url": "https://www.example.com/loadtest"}' > payload.json
```

### Using curl for Simple Load Test

```bash
# Sequential requests
time for i in {1..100}; do
  curl -s -X POST http://localhost:8080/api/shorten \
    -H "Content-Type: application/json" \
    -d "{\"url\": \"https://www.example.com/test$i\"}" > /dev/null
done
```

## Clean Up

After testing:

```bash
# Stop Docker Compose
docker compose down -v

# Or stop individual containers
docker stop gourl_postgres gourl_redis
docker rm gourl_postgres gourl_redis
```

## Automated Test Script

Create a test script to run all tests:

```bash
#!/bin/bash
# test.sh

BASE_URL="http://localhost:8080"

echo "1. Testing health check..."
curl -s $BASE_URL/health | grep -q "healthy" && echo "✓ Health check passed" || echo "✗ Health check failed"

echo -e "\n2. Testing URL creation..."
RESPONSE=$(curl -s -X POST $BASE_URL/api/shorten -H "Content-Type: application/json" -d '{"url":"https://www.google.com"}')
echo $RESPONSE | grep -q "short_code" && echo "✓ URL creation passed" || echo "✗ URL creation failed"

echo -e "\n3. Testing custom code..."
RESPONSE=$(curl -s -X POST $BASE_URL/api/shorten -H "Content-Type: application/json" -d '{"url":"https://www.example.com","custom_code":"test123"}')
echo $RESPONSE | grep -q "test123" && echo "✓ Custom code passed" || echo "✗ Custom code failed"

echo -e "\n4. Testing redirect..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/test123)
[ "$STATUS" = "301" ] && echo "✓ Redirect passed" || echo "✗ Redirect failed"

echo -e "\n5. Testing 404..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/nonexistent)
[ "$STATUS" = "404" ] && echo "✓ 404 test passed" || echo "✗ 404 test failed"

echo -e "\n6. Testing invalid URL..."
RESPONSE=$(curl -s -X POST $BASE_URL/api/shorten -H "Content-Type: application/json" -d '{"url":"not-a-url"}')
echo $RESPONSE | grep -q "error" && echo "✓ Invalid URL test passed" || echo "✗ Invalid URL test failed"

echo -e "\nAll tests completed!"
```

Make it executable and run:
```bash
chmod +x test.sh
./test.sh
```

## Troubleshooting

### Server won't start
- Check if ports 8080, 5432, and 6379 are available
- Verify PostgreSQL and Redis are running
- Check logs: `docker compose logs app`

### Database connection fails
- Ensure PostgreSQL is fully started: `docker compose logs postgres`
- Check connection settings in `.env`
- Verify database is initialized: Check `init.sql` was executed

### Redis connection fails
- Check Redis is running: `docker compose logs redis`
- Verify Redis configuration in `.env`
- Test Redis connection: `docker exec gourl_redis redis-cli ping`

### Rate limiting not working
- Check Redis is running
- Verify `RATE_LIMIT_REQUESTS` and `RATE_LIMIT_WINDOW` in `.env`
- Clear Redis: `docker exec gourl_redis redis-cli flushall`

## Expected Test Results Summary

| Test | Expected HTTP Status | Expected Behavior |
|------|---------------------|-------------------|
| Health check | 200 | Returns `{"status":"healthy"}` |
| Create short URL | 201 | Returns short_code, short_url, long_url |
| Custom code | 201 | Uses provided custom code |
| Redirect | 301 | Redirects to original URL |
| Non-existent code | 404 | Returns error message |
| Invalid URL | 400 | Returns error message |
| Duplicate custom code | 400 | Returns error message |
| Rate limit exceeded | 429 | Returns error message |
| Wrong HTTP method | 405 | Returns error message |
| Missing required field | 400 | Returns error message |
