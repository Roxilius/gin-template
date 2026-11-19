# Untuk Linux/macOS/CI/Git Bash. Di Windows PowerShell gunakan ./scripts/dev.ps1
.DEFAULT_GOAL := help

SWAG_VERSION ?= v1.16.6
BIN_DIR      := bin

.PHONY: help
help: ## Tampilkan daftar perintah
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

## ---------- Development ----------
.PHONY: init run dev build tidy fmt
init: ## Sesuaikan nama project hasil salinan template
	go run ./scripts/initproject

run: ## Jalankan API
	go run ./cmd/api

dev: ## Jalankan API dengan hot reload (air)
	air

build: ## Build binary ke ./bin
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/api ./cmd/api
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/migrate ./cmd/migrate

tidy: ## Rapikan go.mod/go.sum
	go mod tidy

fmt: ## Format kode
	gofmt -w .

## ---------- Quality ----------
.PHONY: test cover lint vet
test: ## Jalankan semua test
	go test ./... -count=1

cover: ## Test + laporan coverage (coverage.html)
	go test ./... -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

vet: ## go vet
	go vet ./...

lint: ## golangci-lint
	golangci-lint run ./...

## ---------- Docs ----------
.PHONY: swag
swag: ## Generate dokumentasi Swagger ke ./docs
	swag init -g cmd/api/main.go -o docs --parseInternal
	swag fmt

## ---------- Database ----------
.PHONY: migrate-up migrate-down migrate-version migrate-create
migrate-up: ## Jalankan semua migration
	go run ./cmd/migrate up

migrate-down: ## Rollback 1 migration (n=N untuk lebih)
	go run ./cmd/migrate down $(or $(n),1)

migrate-version: ## Versi migration saat ini
	go run ./cmd/migrate version

migrate-create: ## Buat file migration baru: make migrate-create name=create_xxx_table
	go run ./cmd/migrate create $(name)

## ---------- Docker ----------
.PHONY: db-up docker-up docker-down docker-logs
db-up: ## Jalankan PostgreSQL saja (untuk development lokal)
	docker compose up -d db

docker-up: ## Jalankan seluruh stack (db + migrate + api)
	docker compose up -d --build

docker-down: ## Hentikan stack
	docker compose down

docker-logs: ## Lihat log api
	docker compose logs -f api

## ---------- Tools ----------
.PHONY: tools
tools: ## Install tool development (air, swag, golangci-lint)
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
