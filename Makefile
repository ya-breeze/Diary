ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: all
all: build test lint
	@echo "🎉 You are good to go!"

.PHONY: build
build:
	@echo "🚀 Building backend..."
	@cd ${ROOT_DIR}/backend/cmd && go build -o ../bin/diary
	@echo "🚀 Building frontend..."
	@cd ${ROOT_DIR}/next-frontend && npm run build
	@echo "✅ Build complete"

.PHONY: run-backend
run-backend:
	@cd ${ROOT_DIR}/backend/cmd && go build -o ../bin/diary
	@GB_USERS=test@test.com:JDJhJDEwJC9sVWJpTlBYVlZvcU9ZNUxIZmhqYi4vUnRuVkJNaEw4MTQ2VUdFSXRDeE9Ib0ZoVkRLR3pl,test:JDJhJDEwJC9sVWJpTlBYVlZvcU9ZNUxIZmhqYi4vUnRuVkJNaEw4MTQ2VUdFSXRDeE9Ib0ZoVkRLR3pl \
	GB_COOKIE_SECURE=false \
	GB_SESSION_SECRET=dev_session_secret_continuous_stable_value \
	GB_JWT_SECRET=dev_jwt_secret_continuous_stable_value \
	GB_DATAPATH=$(ROOT_DIR)diary-data \
	GB_ALLOWEDORIGINS=http://localhost:3000,http://localhost:4200,http://localhost:8080 \
	${ROOT_DIR}/backend/bin/diary server

.PHONY: run-frontend
run-frontend:
	@cd ${ROOT_DIR}/next-frontend && npm run dev


.PHONY: generate_mocks
generate_mocks: generate
	@echo "🚀 Generating mocks..."
	@cd ${ROOT_DIR}/backend && go generate ./...
	@echo "✅ Mocks generated"

.PHONY: generate
generate:
	@echo "🚀 Generating code from OpenAPI spec..."
	@cd ${ROOT_DIR}/backend; \
		go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
			-config ${ROOT_DIR}/api/oapi-codegen-server.yaml \
			${ROOT_DIR}/api/openapi.yaml; \
		go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
			-config ${ROOT_DIR}/api/oapi-codegen-client.yaml \
			${ROOT_DIR}/api/openapi.yaml; \
		goimports -l -w ./pkg/generated/goserver/server.gen.go ./pkg/generated/goclient/client.gen.go; \
		go tool mvdan.cc/gofumpt -l -w ./pkg/generated/goserver/server.gen.go ./pkg/generated/goclient/client.gen.go
	@echo "✅ Generation complete"

.PHONY: validate
validate:
	@echo "🔍 Validating OpenAPI spec..."
	@cd ${ROOT_DIR}/backend; \
		go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
			-config ${ROOT_DIR}/api/oapi-codegen-server.yaml \
			-o /dev/null \
			${ROOT_DIR}/api/openapi.yaml
	@echo "✅ Validation: OpenAPI spec is valid"

.PHONY: lint
lint:
	@echo "🚀 Linting backend..."
	@cd ${ROOT_DIR}/backend; \
		go tool github.com/golangci/golangci-lint/cmd/golangci-lint run; \
		go tool mvdan.cc/gofumpt -l -d .
	@echo "🚀 Linting frontend..."
	@cd ${ROOT_DIR}/next-frontend; \
		npm run lint
	@echo "✅ Lint complete"

.PHONY: test
test:
	@echo "🚀 Running backend tests..."
	@cd ${ROOT_DIR}/backend; \
		go tool github.com/onsi/ginkgo/v2/ginkgo -r
	@echo "✅ Tests complete"

.PHONY: test-e2e
test-e2e:
	@echo "🚀 Running Playwright E2E tests..."
	@cd e2e && BASE_URL=$(or $(BASE_URL),http://192.168.1.54:8885) npx playwright test --reporter=line
	@echo "✅ E2E tests complete"

.PHONY: test-all
test-all: test test-e2e
	@echo "✅ All tests complete"

.PHONY: watch
watch:
	@cd ${ROOT_DIR}/backend && ginkgo watch -r

.PHONE: check-deps
check-deps:
	@echo "🔍 Checking backend dependencies..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed. Please install Go first."; exit 1; }
	@go version
	@echo "🔍 Checking frontend dependencies..."
	@command -v node >/dev/null 2>&1 || { echo "❌ Node.js is required but not installed. Please install Node.js first."; exit 1; }
	@command -v npm >/dev/null 2>&1 || { echo "❌ npm is required but not installed. Please install npm first."; exit 1; }
	@command -v npx >/dev/null 2>&1 || { echo "❌ npx is required but not installed. Please install npx first."; exit 1; }
	@node --version
	@npm --version
	@npx --version
	@echo "🔍 Checking Docker dependencies..."
	@command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed. Please install Docker first."; exit 1; }
	@docker --version
	@echo "✅ Dependencies check complete"

.PHONE: install
install: check-deps
	cd ${ROOT_DIR}/backend && go mod download
	cd ${ROOT_DIR}/next-frontend && npm install

.PHONE: clean
clean:
	@echo "🚀 Cleaning backend..."
	@cd ${ROOT_DIR}/backend && go clean
	@echo "🚀 Cleaning frontend..."
	@cd ${ROOT_DIR}/next-frontend; \
		rm -rf .next/; \
		rm -rf node_modules/; \
		npm cache clean --force
	@echo "✅ Clean complete"

.PHONE: analyze
analyze:
	@echo "📈 Analyzing bundle size..."
	@cd ${ROOT_DIR}/next-frontend; \
		ANALYZE=true npm run build
	@echo "✅ Analysis complete"

# ============================================
# Docker Compose Commands
# ============================================

.PHONY: docker-build
docker-build:
	@echo "🐳 Building Docker images..."
	@docker compose build
	@echo "✅ Docker build complete"

.PHONY: docker-up
docker-up:
	@echo "🐳 Starting Docker containers..."
	@GB_COOKIE_SECURE=false docker compose up -d
	@echo "✅ Docker containers started"
	@echo "📱 Application available at http://localhost"

.PHONY: docker-down
docker-down:
	@echo "🐳 Stopping Docker containers..."
	@docker compose down
	@echo "✅ Docker containers stopped"

.PHONY: docker-logs
docker-logs:
	@docker compose logs -f

.PHONY: docker-restart
docker-restart:
	@echo "🐳 Restarting Docker containers..."
	@GB_COOKIE_SECURE=false docker compose restart
	@echo "✅ Docker containers restarted"

.PHONY: docker-clean
docker-clean:
	@echo "🐳 Cleaning Docker containers and volumes..."
	@docker compose down -v
	@echo "✅ Docker cleanup complete"

.PHONY: compose
compose: docker-build docker-up
	@echo "🎉 Docker Compose deployment complete!"
	@echo "📱 Access the application at http://localhost"
	@echo "📚 See DOCKER.md for more information"
