# Personal Diary Frontend - Makefile
# Common development tasks for Angular application

.PHONY: help install dev build test e2e lint clean deploy setup docker-build docker-run

# Default target
help:
	@echo "Personal Diary Frontend - Available Commands:"
	@echo ""
	@echo "Setup and Installation:"
	@echo "  make setup     - Initial project setup (install Node.js dependencies)"
	@echo "  make install   - Install/update npm dependencies"
	@echo ""
	@echo "Development:"
	@echo "  make dev       - Start development server with hot reload"
	@echo "  make build     - Build application for production"
	@echo "  make build-dev - Build application for development"
	@echo ""
	@echo "Testing:"
	@echo "  make test      - Run unit tests"
	@echo "  make test-watch - Run unit tests in watch mode"
	@echo "  make e2e       - Run end-to-end tests"
	@echo "  make coverage  - Generate test coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint      - Run ESLint and check code style"
	@echo "  make lint-fix  - Run ESLint and auto-fix issues"
	@echo "  make format    - Format code with Prettier"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean     - Clean build artifacts and node_modules"
	@echo "  make analyze   - Analyze bundle size"
	@echo "  make serve-prod - Serve production build locally"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run application in Docker container"
	@echo ""
	@echo "Deployment:"
	@echo "  make deploy-staging - Deploy to staging environment"
	@echo "  make deploy-prod    - Deploy to production environment"

# Setup and Installation
setup: install
	@echo "✅ Project setup complete!"

install:
	@echo "📦 Installing dependencies..."
	npm install

# Development
dev:
	@echo "🚀 Starting development server..."
	npm run start

build:
	@echo "🏗️  Building for production..."
	npm run build

build-dev:
	@echo "🏗️  Building for development..."
	npm run build -- --configuration development

# Testing
test:
	@echo "🧪 Running unit tests..."
	npm run test -- --watch=false --browsers=ChromeHeadless

test-watch:
	@echo "🧪 Running unit tests in watch mode..."
	npm run test

e2e:
	@echo "🔍 Running end-to-end tests..."
	npm run e2e

coverage:
	@echo "📊 Generating test coverage report..."
	npm run test -- --watch=false --browsers=ChromeHeadless --code-coverage

# Code Quality
lint:
	@echo "🔍 Running ESLint..."
	npm run lint

lint-fix:
	@echo "🔧 Running ESLint with auto-fix..."
	npm run lint -- --fix

format:
	@echo "💅 Formatting code with Prettier..."
	npx prettier --write "src/**/*.{ts,html,css,scss,json}"

# Utilities
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf dist/
	rm -rf node_modules/
	rm -rf coverage/
	npm cache clean --force

analyze:
	@echo "📈 Analyzing bundle size..."
	npm run build -- --stats-json
	npx webpack-bundle-analyzer dist/stats.json

serve-prod:
	@echo "🌐 Serving production build..."
	npx http-server dist/ -p 8080 -c-1

# Docker
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t diary-frontend .

docker-run:
	@echo "🐳 Running Docker container..."
	docker run -p 4200:80 diary-frontend

# Deployment (customize these based on your deployment strategy)
deploy-staging:
	@echo "🚀 Deploying to staging..."
	npm run build -- --configuration staging
	# Add your staging deployment commands here

deploy-prod:
	@echo "🚀 Deploying to production..."
	npm run build -- --configuration production
	# Add your production deployment commands here

# Development helpers
generate-component:
	@read -p "Enter component name: " name; \
	ng generate component $$name

generate-service:
	@read -p "Enter service name: " name; \
	ng generate service $$name

generate-module:
	@read -p "Enter module name: " name; \
	ng generate module $$name

# API Documentation
api-docs:
	@echo "📚 Opening API documentation..."
	@if command -v xdg-open > /dev/null; then \
		xdg-open api/openapi.yaml; \
	elif command -v open > /dev/null; then \
		open api/openapi.yaml; \
	else \
		echo "Please open api/openapi.yaml manually"; \
	fi

# Check if Node.js and npm are installed
check-deps:
	@echo "🔍 Checking dependencies..."
	@command -v node >/dev/null 2>&1 || { echo "❌ Node.js is required but not installed. Please install Node.js first."; exit 1; }
	@command -v npm >/dev/null 2>&1 || { echo "❌ npm is required but not installed. Please install npm first."; exit 1; }
	@echo "✅ Node.js and npm are installed"
	@node --version
	@npm --version
