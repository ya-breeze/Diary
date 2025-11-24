ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: all
all: build test validate lint
	@echo "🎉 You are good to go!"

.PHONY: build
build:
	@echo "🚀 Building backend..."
	@cd ${ROOT_DIR}/backend/cmd && go build -o ../bin/diary
	@echo "🚀 Building frontend..."
	@cd ${ROOT_DIR}/frontend && npm run build
	@echo "🚀 Building Docker image..."
	@docker build -t diary .
	@echo "✅ Build complete"

.PHONY: run-backend
run-backend: build
	@GB_USERS=test@test.com:JDJhJDEwJC9sVWJpTlBYVlZvcU9ZNUxIZmhqYi4vUnRuVkJNaEw4MTQ2VUdFSXRDeE9Ib0ZoVkRLR3pl,test:JDJhJDEwJC9sVWJpTlBYVlZvcU9ZNUxIZmhqYi4vUnRuVkJNaEw4MTQ2VUdFSXRDeE9Ib0ZoVkRLR3pl \
	GB_DISABLEIMPORTERS=true \
	GB_DBPATH=$(ROOT_DIR)diary.db \
	GB_ASSETPATH=$(ROOT_DIR)/diary-assets \
	${ROOT_DIR}/backend/bin/diary server

.PHONY: run-frontend
run-frontend: build
	@cd ${ROOT_DIR}/frontend && npm run start

.PHONY: replace-templates
replace-templates:
	@cd ${ROOT_DIR}/backend
	@rm -rf pkg/generated/templates/goclient pkg/generated/templates/goserver
	@mkdir -p pkg/generated/templates/goclient pkg/generated/templates/goserver
	@docker run --rm -u 1000 -v ${HOST_PWD}:/local \
		openapitools/openapi-generator-cli author template -g go \
		-o /local/pkg/generated/templates/goclient
	@docker run --rm -u 1000 -v ${HOST_PWD}:/local \
		openapitools/openapi-generator-cli author template -g go-server \
		-o /local/pkg/generated/templates/goserver

.PHONY: generate_mocks
generate_mocks: generate
	@echo "🚀 Generating mocks..."
	@cd ${ROOT_DIR}/backend
	@go generate ./...
	@echo "✅ Mocks generated"

.PHONY: generate
generate:
	@echo "🚀 Generating code from OpenAPI spec..."
	@cd ${ROOT_DIR}/backend
	# Golang client and server
	@rm -rf pkg/generated/goclient pkg/generated/goserver pkg/generated/angular
	@mkdir -p pkg/generated/goclient pkg/generated/goserver
	@docker run --rm -u 1000 -v ${HOST_PWD}:/local \
		openapitools/openapi-generator-cli generate \
		-i /local/api/openapi.yaml \
		-g go \
		-t /local/pkg/generated/templates/goclient \
		-o /local/pkg/generated/goclient \
		--additional-properties=packageName=goclient,withGoMod=false
	@rm -rf \
		pkg/generated/goclient/api \
		pkg/generated/goclient/.gitignore \
		pkg/generated/goclient/.openapi-generator-ignore \
		pkg/generated/goclient/.travis.yml \
		pkg/generated/goclient/*.sh \
		pkg/generated/goclient/go.* \
		pkg/generated/goclient/test
	@docker run --rm -u 1000 -v ${HOST_PWD}:/local \
		openapitools/openapi-generator-cli generate \
		-i /local/api/openapi.yaml \
		-g go-server \
		-t /local/pkg/generated/templates/goserver \
		-o /local/pkg/generated/goserver \
		--additional-properties=packageName=goserver,featureCORS=true,hideGenerationTimestamp=true
	@rm -rf \
		pkg/generated/goserver/api \
		pkg/generated/goserver/.openapi-generator-ignore \
		pkg/generated/goserver/Dockerfile \
		pkg/generated/goserver/go.*
	@mv -f pkg/generated/goserver/go/* pkg/generated/goserver
	@rm -rf pkg/generated/goserver/go
	@goimports -l -w ./pkg/generated/
	@go tool mvdan.cc/gofumpt -l -w ./pkg/generated/

	@echo "✅ Generation complete"

.PHONY: validate
validate:
	@cd ${ROOT_DIR}/backend
	@docker run --rm -v ${HOST_PWD}:/local openapitools/openapi-generator-cli validate -i /local/api/openapi.yaml
	@echo "✅ Validation complete"

.PHONY: lint
lint:
	@echo "🚀 Linting backend..."
	@cd ${ROOT_DIR}/backend
	@go tool github.com/golangci/golangci-lint/cmd/golangci-lint run
	@go tool mvdan.cc/gofumpt -l -d .
	@echo "🚀 Linting frontend..."
	@cd ${ROOT_DIR}/frontend
	@npx prettier --write "src/**/*.{ts,html,css,scss,json}"
	@npm run lint -- --fix
	@echo "✅ Lint complete"

.PHONY: test
test:
	@echo "🚀 Running backend tests..."
	@cd ${ROOT_DIR}/backend
	@go tool github.com/onsi/ginkgo/v2/ginkgo -r
	@echo "🚀 Running frontend tests..."
	@cd ${ROOT_DIR}/frontend
	CHROME_BIN=chromium npm run test -- --watch=false --browsers=ChromeHeadless
	@echo "✅ Tests complete"

.PHONY: watch
watch:
	@cd ${ROOT_DIR}/backend
	@ginkgo watch -r

.PHONE: compose
compose:
	@docker-compose up --build

.PHONE: check-deps
check-deps:
	@echo "🔍 Checking backend dependencies..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed. Please install Go first."; exit 1; }
	@go version
	@echo "🔍 Checking frontend dependencies..."
	@command -v node >/dev/null 2>&1 || { echo "❌ Node.js is required but not installed. Please install Node.js first."; exit 1; }
	@command -v npm >/dev/null 2>&1 || { echo "❌ npm is required but not installed. Please install npm first."; exit 1; }
	@command -v npx >/dev/null 2>&1 || { echo "❌ npx is required but not installed. Please install npx first."; exit 1; }
	@echo "✅ Node.js and npm are installed"
	@node --version
	@npm --version
	@npx --version

.PHONE: install
install: check-deps
	@cd ${ROOT_DIR}/backend
	@go mod tidy
	@go mod download
	@cd ${ROOT_DIR}/frontend
	@npm install

.PHONE: clean
clean:
	@echo "🚀 Cleaning backend..."
	@cd ${ROOT_DIR}/backend
	@go clean
	@echo "🚀 Cleaning frontend..."
	@cd ${ROOT_DIR}/frontend
	@rm -rf dist/
	@rm -rf node_modules/
	@rm -rf coverage/
	@npm cache clean --force
	@echo "✅ Clean complete"

.PHONE: analyze
analyze:
	@echo "📈 Analyzing bundle size..."
	@cd ${ROOT_DIR}/frontend
	@npm run build -- --stats-json
	@npx webpack-bundle-analyzer dist/stats.json
	@echo "✅ Analysis complete"
