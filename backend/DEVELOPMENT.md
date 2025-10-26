# Development Guide

## Using Air for Live Reloading

[Air](https://github.com/air-verse/air) provides automatic code reloading during development.

### Local Development (Recommended)

```bash
cd backend
air
```

This uses your local Go installation and is faster than Docker.

### Docker Development

For development with Docker, the setup uses **Docker volumes** to sync files:

```yaml
volumes:
  - ./backend:/app        # Your host files appear in /app inside container
  - /app/tmp             # Exclude Air's tmp directory
  - go-mod-cache:/go/pkg/mod  # Cache Go modules for speed
```

**How it works:**
1. Your local `./backend` directory is **mounted** into `/app` in the container
2. When you edit files on your host, Docker reflects the changes in the container
3. Air watches the mounted files and rebuilds on changes

**To start:**
```bash
# Start all services (postgres, redis, backend)
docker-compose -f docker-compose.dev.yml up

# Or in background
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f backend
```

## Troubleshooting

### File Changes Not Detected

If file changes aren't being detected in Docker:

1. **Check volume mounting:**
   ```bash
   # Enter the container
   docker exec -it chat-ecommerce-backend-dev sh
   
   # Check if files are there
   ls -la /app
   ```

2. **Enable verbose logging:**
   Air configuration already has polling enabled. If issues persist:
   
   ```bash
   # Run Air with debug mode
   air -d
   ```

3. **Restart container:**
   ```bash
   docker-compose -f docker-compose.dev.yml restart backend
   ```

### Performance Issues

Docker volumes can be slower than native on macOS/Windows. Options:

1. **Use local development** (fastest):
   ```bash
   cd backend
   air
   ```

2. **Use VS Code Remote-Containers** extension for better integration

3. **WSL2 on Windows** provides better Docker performance

## Configuration

Air configuration is in `.air.toml`. You can customize:

- **Files to watch**: Add extensions to `include_ext`
- **Directories to ignore**: Add to `exclude_dir`
- **Build command**: Modify `cmd` in `[build]` section
- **Polling interval**: Adjust `poll_interval` (milliseconds)

## Environment Variables

Create a `.env` file in the project root:

```bash
cp env.example .env
```

Key variables for development:
- `DB_HOST`: Database host (use `postgres` in Docker, `localhost` locally)
- `REDIS_HOST`: Redis host (use `redis` in Docker, `localhost` locally)
- `PORT`: Server port (default: 8080)

## Database Setup

The database is automatically migrated and seeded when the backend starts.

To reset:
```bash
docker-compose down -v  # Removes volumes
docker-compose up -d postgres
```

## Testing Changes

```bash
# Run all tests
go test ./...

# Run specific test
go test ./internal/handlers/

# Watch and test
gotestsum --watch
```

## Helpful Commands

```bash
# View Air logs
air -d  # Debug mode with all logs

# Rebuild Docker image
docker-compose -f docker-compose.dev.yml build backend

# Clean up
docker-compose -f docker-compose.dev.yml down
rm -rf backend/tmp
```
