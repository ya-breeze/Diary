# Multi-stage build for Go backend
FROM golang:1.24-alpine AS builder

# Install build dependencies (including gcc for CGO/SQLite)
RUN apk add --no-cache git make gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy backend source code
COPY backend/ ./
COPY api/ ../api/

# Build the application with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o /app/bin/diary ./cmd/main.go

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies (including sqlite libs for runtime)
RUN apk --no-cache add ca-certificates wget tzdata sqlite-libs

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/diary /app/diary

# Create data directories
RUN mkdir -p /data/diary-assets

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK NONE

# Run the application
CMD ["/app/diary", "server"]

