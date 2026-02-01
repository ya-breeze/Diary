# Docker Deployment Guide

This guide explains how to deploy the Diary application using Docker Compose with Nginx as a reverse proxy.

## Architecture

The Docker Compose setup consists of three containers:

1. **Backend Container**: Go API server running on port 8080 (internal)
2. **Frontend Container**: Next.js server running on port 3000 (internal)
3. **Nginx Proxy Container**: Reverse proxy listening on ports 80/443 (external)

### Request Flow

```
User Request
    ↓
Nginx Proxy (Port 80/443)
    ↓
    ├─→ /v1/* → Backend Container (Go API)
    ├─→ /web/* → Backend Container (Web Interface)
    └─→ /* → Frontend Container (Next.js App)
```

## Quick Start

### 1. Prerequisites

- Docker 20.10+
- Docker Compose 2.0+

### 2. Configuration

Copy the example environment file and customize it:

```bash
cp .env.example .env
```

Edit `.env` to configure your deployment:

```bash
# Example: Change allowed origins for production
GB_ALLOWEDORIGINS=https://yourdomain.com,https://www.yourdomain.com

# Example: Use external volume for data persistence
DIARY_DATA_PATH=/var/lib/diary-data

# Example: Change exposed ports
NGINX_HTTP_PORT=8080
NGINX_HTTPS_PORT=8443
```

### 3. Build and Run

```bash
# Build and start all containers
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop containers
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### 4. Access the Application

- **Web Interface**: http://localhost
- **API**: http://localhost/v1
- **Health Check**: http://localhost/health

Default credentials:

- **Email**: test@test.com
- **Password**: test

## Environment Variables

### Frontend Configuration

| Variable              | Description                       | Default |
| --------------------- | --------------------------------- | ------- |
| `NEXT_PUBLIC_API_URL` | API URL for frontend (browser side)| /api    |

**Note**: The frontend uses `NEXT_PUBLIC_API_URL` to determine the API endpoint. Use a relative path (`/api`) when behind the Nginx proxy, or a full URL (`http://localhost:8080/v1`) for direct access.

### Backend Configuration

| Variable                 | Description                                  | Default              |
| ------------------------ | -------------------------------------------- | -------------------- |
| `GB_USERS`               | User credentials (email:bcrypt_hash)         | test@test.com:...    |
| `GB_DATAPATH`            | Data directory path (inside container)       | /data                |
| `GB_ALLOWEDORIGINS`      | CORS allowed origins (comma-separated)       | http://localhost,... |
| `GB_JWTSECRET`           | JWT signing secret (auto-generated if empty) | -                    |
| `GB_ISSUER`              | JWT issuer identifier                        | diary                |
| `GB_PORT`                | Backend port (inside container)              | 8080                 |
| `GB_MAXPERFILESIZEMB`    | Max file size (MB)                           | 200                  |
| `GB_MAXBATCHFILES`       | Max files per batch                          | 100                  |
| `GB_MAXBATCHTOTALSIZEMB` | Max batch size (MB)                          | 1000                 |

### Volume Configuration

| Variable          | Description                    | Default      |
| ----------------- | ------------------------------ | ------------ |
| `DIARY_DATA_PATH` | Host path for data persistence | ./diary-data |

### Nginx Configuration

| Variable           | Description        | Default |
| ------------------ | ------------------ | ------- |
| `NGINX_HTTP_PORT`  | HTTP port on host  | 80      |
| `NGINX_HTTPS_PORT` | HTTPS port on host | 443     |

## Data Persistence

All application data (database and assets) is stored in a Docker volume mapped to the host filesystem.

### Default Location

By default, data is stored in `./diary-data` relative to the docker-compose.yml file.

### Custom Location

To use a different location, set the `DIARY_DATA_PATH` environment variable:

```bash
# In .env file
DIARY_DATA_PATH=/var/lib/diary-data
```

Or override when running:

```bash
DIARY_DATA_PATH=/custom/path docker-compose up -d
```

### Backup

To backup your data:

```bash
# Stop the application
docker-compose down

# Backup the data directory
tar -czf diary-backup-$(date +%Y%m%d).tar.gz diary-data/

# Restart the application
docker-compose up -d
```

### Restore

To restore from backup:

```bash
# Stop the application
docker-compose down

# Restore the data directory
tar -xzf diary-backup-YYYYMMDD.tar.gz

# Restart the application
docker-compose up -d
```

## SSL/HTTPS Configuration

To enable HTTPS:

### 1. Obtain SSL Certificates

Place your SSL certificates in `nginx/ssl/`:

```bash
mkdir -p nginx/ssl
cp /path/to/cert.pem nginx/ssl/
cp /path/to/key.pem nginx/ssl/
```

### 2. Enable HTTPS in Nginx Configuration

Edit `nginx/conf.d/default.conf` and uncomment the HTTPS server block.

### 3. Update docker-compose.yml

Uncomment the SSL volume mount in the nginx service:

```yaml
volumes:
  - ./nginx/ssl:/etc/nginx/ssl:ro
```

### 4. Restart Nginx

```bash
docker-compose restart nginx
```

## Production Deployment

### Security Checklist

- [ ] Change default user credentials
- [ ] Set strong `GB_JWTSECRET`
- [ ] Configure proper `GB_ALLOWEDORIGINS`
- [ ] Enable HTTPS with valid SSL certificates
- [ ] Use external volume for data persistence
- [ ] Configure firewall rules
- [ ] Set up regular backups
- [ ] Review and adjust file upload limits
- [ ] Enable log rotation

### Recommended .env for Production

```bash
GB_USERS=admin@yourdomain.com:YOUR_BCRYPT_HASH
GB_ALLOWEDORIGINS=https://yourdomain.com
GB_JWTSECRET=YOUR_STRONG_SECRET_KEY_HERE
DIARY_DATA_PATH=/var/lib/diary-data
NGINX_HTTP_PORT=80
NGINX_HTTPS_PORT=443
```

## Troubleshooting

### View Logs

```bash
# All containers
docker-compose logs -f

# Specific container
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f nginx
```

### Check Container Status

```bash
docker-compose ps
```

### Restart Containers

```bash
# Restart all
docker-compose restart

# Restart specific container
docker-compose restart backend
```

### Rebuild Containers

```bash
# Rebuild and restart
docker-compose up -d --build

# Rebuild specific container
docker-compose up -d --build backend
```

### Access Container Shell

```bash
# Backend
docker-compose exec backend sh

# Frontend
docker-compose exec frontend sh

# Nginx
docker-compose exec nginx sh
```

### Common Issues

#### CORS Errors

Ensure `GB_ALLOWEDORIGINS` includes your frontend URL:

```bash
GB_ALLOWEDORIGINS=http://localhost,http://localhost:80
```

#### Permission Denied on Data Directory

Ensure the data directory has correct permissions:

```bash
sudo chown -R 1000:1000 diary-data/
```

#### Port Already in Use

Change the exposed ports in `.env`:

```bash
NGINX_HTTP_PORT=8080
NGINX_HTTPS_PORT=8443
```

## Monitoring

### Health Checks

All containers have health checks configured:

```bash
# Check health status
docker-compose ps
```

### Resource Usage

```bash
# View resource usage
docker stats
```

## Scaling

To run multiple backend instances:

```bash
docker-compose up -d --scale backend=3
```

Note: You'll need to configure Nginx load balancing in `nginx/conf.d/default.conf`.

## Maintenance

### Update Application

```bash
# Pull latest changes
git pull

# Rebuild and restart
docker-compose up -d --build
```

### Clean Up

```bash
# Remove stopped containers
docker-compose down

# Remove all containers, networks, and volumes
docker-compose down -v

# Remove unused images
docker image prune -a
```
