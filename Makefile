# Personal Diary Frontend - Makefile
# Common development tasks for Angular application

.PHONY: help install dev build test e2e lint clean deploy setup docker-build docker-run

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

# Check if Node.js and npm are installed
check-deps:
	@echo "🔍 Checking dependencies..."
	@command -v node >/dev/null 2>&1 || { echo "❌ Node.js is required but not installed. Please install Node.js first."; exit 1; }
	@command -v npm >/dev/null 2>&1 || { echo "❌ npm is required but not installed. Please install npm first."; exit 1; }
	@echo "✅ Node.js and npm are installed"
	@node --version
	@npm --version
