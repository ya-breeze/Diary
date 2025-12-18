#!/bin/sh
set -e

# Start nginx
echo "Starting nginx..."

# Substitute environment variables in config template
envsubst < /usr/share/nginx/html/assets/config.template.json > /usr/share/nginx/html/assets/config.json

exec nginx -g "daemon off;"

