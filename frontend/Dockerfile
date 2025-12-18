# Multi-stage build for Angular frontend
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY frontend/package*.json ./

# Install dependencies (including devDependencies needed for build)
RUN npm ci && \
  npm cache clean --force

# Copy source code
COPY frontend/ ./

# Build the application
RUN npm run build

# Final stage - Nginx to serve static files
FROM nginx:alpine

# Install gettext for envsubst
RUN apk add --no-cache gettext

# Remove default nginx config
RUN rm -rf /usr/share/nginx/html/*

# Copy built Angular app
COPY --from=builder /app/dist/Diary.FE/browser /usr/share/nginx/html

# Copy custom nginx config for SPA routing
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Copy entrypoint script
COPY frontend/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Expose port
EXPOSE 4200

# Health check
HEALTHCHECK NONE

# Use custom entrypoint to generate config at runtime
ENTRYPOINT ["/docker-entrypoint.sh"]

