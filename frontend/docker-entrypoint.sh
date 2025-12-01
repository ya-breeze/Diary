#!/bin/sh
set -e

# Generate config.json from environment variables
echo "Generating runtime configuration..."

# Default API URL if not set
API_URL="${API_URL:-/v1}"

# Create assets directory if it doesn't exist
mkdir -p /usr/share/nginx/html/assets

# Create config.json from template
cat > /usr/share/nginx/html/assets/config.json <<EOF
{
  "apiUrl": "${API_URL}"
}
EOF

echo "Runtime configuration generated:"
cat /usr/share/nginx/html/assets/config.json

# Start nginx
echo "Starting nginx..."
exec nginx -g "daemon off;"

