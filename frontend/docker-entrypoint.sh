#!/bin/sh
set -e

# Start nginx
echo "Starting nginx..."

exec nginx -g "daemon off;"

