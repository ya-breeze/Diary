# Nginx Configuration

This directory contains the Nginx reverse proxy configuration for the Diary application.

## Files

- **nginx.conf** - Main Nginx configuration (global settings)
- **conf.d/default.conf** - Reverse proxy routing configuration

## Routing Rules

### API Routes (`/v1/*`)
All requests to `/v1/*` are proxied to the backend Go API server.

**Example:**
- `http://localhost/v1/items` → `http://backend:8080/v1/items`
- `http://localhost/v1/authorize` → `http://backend:8080/v1/authorize`

### Web Interface Routes (`/web/*`)
All requests to `/web/*` are proxied to the backend web interface.

**Example:**
- `http://localhost/web/login` → `http://backend:8080/web/login`

### Frontend Routes (`/*`)
All other requests are proxied to the frontend Angular application.

**Example:**
- `http://localhost/` → `http://frontend:4200/`
- `http://localhost/diary` → `http://frontend:4200/diary`

## HTTPS Configuration

To enable HTTPS:

### 1. Obtain SSL Certificates

Place your SSL certificates in the `ssl/` directory:

```bash
mkdir -p nginx/ssl
cp /path/to/your/cert.pem nginx/ssl/
cp /path/to/your/key.pem nginx/ssl/
```

### 2. Update docker-compose.yml

Uncomment the SSL volume mount in the nginx service:

```yaml
volumes:
  - ./nginx/ssl:/etc/nginx/ssl:ro
```

### 3. Enable HTTPS Server Block

Edit `conf.d/default.conf` and uncomment the HTTPS server block (lines starting with `# server {`).

### 4. Restart Nginx

```bash
docker-compose restart nginx
```

## Custom Configuration

### Adding Custom Headers

Edit `conf.d/default.conf` and add headers in the server block:

```nginx
add_header X-Custom-Header "value" always;
```

### Changing Upload Limits

Edit `nginx.conf` and modify:

```nginx
client_max_body_size 1000M;  # Change to desired size
```

### Adding Rate Limiting

Add to `nginx.conf` in the `http` block:

```nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
```

Then in `conf.d/default.conf` in the location block:

```nginx
location /v1/ {
    limit_req zone=api burst=20;
    # ... rest of configuration
}
```

### Adding IP Whitelisting

Add to the location block in `conf.d/default.conf`:

```nginx
location /v1/ {
    allow 192.168.1.0/24;
    deny all;
    # ... rest of configuration
}
```

## Load Balancing

To add load balancing for multiple backend instances:

### 1. Update upstream in conf.d/default.conf

```nginx
upstream backend {
    least_conn;  # or ip_hash, or round-robin (default)
    server backend:8080;
    server backend2:8080;
    server backend3:8080;
}
```

### 2. Scale backend in docker-compose

```bash
docker-compose up -d --scale backend=3
```

## Monitoring

### Access Logs

View access logs:

```bash
docker-compose exec nginx tail -f /var/log/nginx/access.log
```

### Error Logs

View error logs:

```bash
docker-compose exec nginx tail -f /var/log/nginx/error.log
```

### Test Configuration

Test Nginx configuration without restarting:

```bash
docker-compose exec nginx nginx -t
```

### Reload Configuration

Reload Nginx configuration without downtime:

```bash
docker-compose exec nginx nginx -s reload
```

## Troubleshooting

### 502 Bad Gateway

This usually means the backend or frontend container is not running or not accessible.

**Check:**
1. Container status: `docker-compose ps`
2. Backend health: `docker-compose exec backend wget -O- http://localhost:8080/v1/user`
3. Frontend health: `docker-compose exec frontend wget -O- http://localhost:4200/`

### 413 Request Entity Too Large

The uploaded file is too large.

**Solution:** Increase `client_max_body_size` in `nginx.conf`

### CORS Errors

Ensure the backend `GB_ALLOWEDORIGINS` includes the Nginx proxy URL.

**Example:**
```bash
GB_ALLOWEDORIGINS=http://localhost,http://localhost:80
```

### Connection Timeout

Increase timeout values in `conf.d/default.conf`:

```nginx
proxy_connect_timeout 120s;
proxy_send_timeout 120s;
proxy_read_timeout 120s;
```

## Security Best Practices

- ✅ Use HTTPS in production
- ✅ Keep Nginx updated
- ✅ Use strong SSL/TLS configuration
- ✅ Enable security headers (already configured)
- ✅ Implement rate limiting for APIs
- ✅ Use IP whitelisting for admin endpoints
- ✅ Regularly review access logs
- ✅ Keep SSL certificates up to date

## Performance Tuning

### Enable Caching

Add to `conf.d/default.conf`:

```nginx
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=api_cache:10m max_size=1g inactive=60m;

location /v1/ {
    proxy_cache api_cache;
    proxy_cache_valid 200 10m;
    proxy_cache_key "$scheme$request_method$host$request_uri";
    # ... rest of configuration
}
```

### Increase Worker Connections

Edit `nginx.conf`:

```nginx
events {
    worker_connections 2048;  # Increase from 1024
}
```

### Enable HTTP/2

Already enabled in the HTTPS server block:

```nginx
listen 443 ssl http2;
```

