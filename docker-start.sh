#!/bin/bash

# Diary Application - Docker Quick Start Script
# This script helps you quickly deploy the Diary application using Docker Compose

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored message
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Print header
print_header() {
    echo ""
    print_message "$BLUE" "=========================================="
    print_message "$BLUE" "  Diary Application - Docker Deployment"
    print_message "$BLUE" "=========================================="
    echo ""
}

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_message "$RED" "❌ Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_message "$RED" "❌ Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_message "$GREEN" "✅ Docker and Docker Compose are installed"
}

# Create .env file if it doesn't exist
setup_env() {
    if [ ! -f .env ]; then
        print_message "$YELLOW" "⚙️  Creating .env file from .env.example..."
        cp .env.example .env
        print_message "$GREEN" "✅ .env file created"
        print_message "$YELLOW" "📝 Please review and customize .env file if needed"
    else
        print_message "$GREEN" "✅ .env file already exists"
    fi
}

# Create data directory
setup_data_dir() {
    local data_dir="./diary-data"
    
    if [ ! -d "$data_dir" ]; then
        print_message "$YELLOW" "📁 Creating data directory..."
        mkdir -p "$data_dir"
        print_message "$GREEN" "✅ Data directory created: $data_dir"
    else
        print_message "$GREEN" "✅ Data directory already exists: $data_dir"
    fi
}

# Build Docker images
build_images() {
    print_message "$BLUE" "🐳 Building Docker images..."
    docker-compose build
    print_message "$GREEN" "✅ Docker images built successfully"
}

# Start containers
start_containers() {
    print_message "$BLUE" "🚀 Starting Docker containers..."
    docker-compose up -d
    print_message "$GREEN" "✅ Docker containers started"
}

# Show status
show_status() {
    echo ""
    print_message "$BLUE" "📊 Container Status:"
    docker-compose ps
    echo ""
}

# Show access information
show_access_info() {
    print_message "$GREEN" "=========================================="
    print_message "$GREEN" "  🎉 Deployment Complete!"
    print_message "$GREEN" "=========================================="
    echo ""
    print_message "$BLUE" "📱 Access the application:"
    echo "   Web Interface: http://localhost"
    echo "   API: http://localhost/v1"
    echo "   Health Check: http://localhost/health"
    echo ""
    print_message "$BLUE" "🔐 Default Credentials:"
    echo "   Email: test@test.com"
    echo "   Password: test"
    echo ""
    print_message "$YELLOW" "📚 Useful Commands:"
    echo "   View logs:        docker-compose logs -f"
    echo "   Stop containers:  docker-compose down"
    echo "   Restart:          docker-compose restart"
    echo "   Rebuild:          docker-compose up -d --build"
    echo ""
    print_message "$YELLOW" "📖 For more information, see DOCKER.md"
    echo ""
}

# Main function
main() {
    print_header
    
    # Check prerequisites
    check_docker
    
    # Setup
    setup_env
    setup_data_dir
    
    # Build and start
    build_images
    start_containers
    
    # Wait a bit for containers to start
    print_message "$YELLOW" "⏳ Waiting for containers to start..."
    sleep 5
    
    # Show status and access info
    show_status
    show_access_info
}

# Run main function
main

