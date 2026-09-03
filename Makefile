GO ?= go

APP_NAME := notification
CMD := ./cmd/notification
BIN_DIR := bin
BINARY := $(BIN_DIR)/$(APP_NAME)
TOOLS_DIR := $(BIN_DIR)/tools
GOOSE_VERSION := v3.26.0
GOOSE := $(TOOLS_DIR)/goose
MIGRATIONS_DIR := ./migrations

.DEFAULT_GOAL := build

.PHONY: build deps fmt vet test coverage run check check-database-url migrate-up migrate-down migrate-status

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BINARY) $(CMD)

deps:
	$(GO) mod download

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

run:
	$(GO) run $(CMD)

check: fmt vet test build

$(GOOSE):
	@mkdir -p $(TOOLS_DIR)
	GOBIN=$(abspath $(TOOLS_DIR)) $(GO) install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)

check-database-url:
	@test -n "$$DATABASE_URL" || { echo "DATABASE_URL is required" >&2; exit 1; }

migrate-up: check-database-url $(GOOSE)
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" up

migrate-down: check-database-url $(GOOSE)
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" down

migrate-status: check-database-url $(GOOSE)
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" status
