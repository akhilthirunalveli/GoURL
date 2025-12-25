# Quick Start Guide

Get the GoURL URL shortener up and running in 5 minutes!

## Prerequisites

- Docker and Docker Compose installed on your system
- curl (for testing)

## Step 1: Clone the Repository

```bash
git clone https://github.com/akhilthirunalveli/GoURL.git
cd GoURL
```

## Step 2: Start the Services

```bash
docker compose up -d
```

This will start:
- PostgreSQL database on port 5432
- Redis cache on port 6379
- GoURL application on port 8080

Wait a few seconds for services to initialize.

## Step 3: Verify the Service

```bash
curl http://localhost:8080/health
```

You should see:
```json
{"status":"healthy"}
```

## Step 4: Create Your First Short URL

```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com"}'
```

You'll get a response like:
```json
{
  "short_code": "r1ZBjJ",
  "short_url": "http://localhost:8080/r1ZBjJ",
  "long_url": "https://github.com"
}
```

## Step 5: Test the Redirect

```bash
curl -L http://localhost:8080/r1ZBjJ
```

You'll be redirected to https://github.com!

## Step 6: Create a Custom Short URL (Optional)

```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.google.com", "custom_code": "search"}'
```

Now you can use: http://localhost:8080/search

## Common Commands

### View Logs
```bash
# All services
docker compose logs

# Just the app
docker compose logs app

# Follow logs
docker compose logs -f app
```

### Stop the Services
```bash
docker compose down
```

### Stop and Remove Data
```bash
docker compose down -v
```

### Restart the Services
```bash
docker compose restart
```

## Configuration

To customize settings, create a `.env` file:

```bash
cp .env.example .env
# Edit .env with your preferred settings
```

Then restart:
```bash
docker compose down
docker compose up -d
```

## What's Next?

- Read the [README.md](README.md) for detailed documentation
- Check [API.md](API.md) for complete API reference
- See [TESTING.md](TESTING.md) for comprehensive testing guide
- Review [IMPLEMENTATION.md](IMPLEMENTATION.md) for architecture details

## Troubleshooting

### Services won't start

Check if ports are available:
```bash
# Check if ports are in use
lsof -i :8080  # Application
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis
```

### Database connection errors

Wait a bit longer for PostgreSQL to initialize:
```bash
docker compose logs postgres
```

### Redis connection errors

Check Redis is running:
```bash
docker compose logs redis
```

### Reset everything

```bash
docker compose down -v
docker compose up -d --build
```

## Need Help?

- Check the logs: `docker compose logs`
- Verify health: `curl http://localhost:8080/health`
- Read full docs: [README.md](README.md)

## Production Deployment

For production use:
1. Use HTTPS (set up reverse proxy with SSL)
2. Change default passwords
3. Configure proper domains
4. Set up monitoring
5. Configure backups

See [README.md](README.md) for production considerations.

---

**That's it!** You now have a fully functional URL shortener running locally. 🎉
