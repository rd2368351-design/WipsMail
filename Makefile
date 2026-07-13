SHELL := /bin/bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

APP_NAME          := wispmail
CTL_NAME          := wispctl
WORKER_NAME       := wispmail-worker
BIN_DIR           := ./bin
GO_VERSION        := 1.26.5
LDFLAGS           := -ldflags="-w -s -X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev) -X main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: init
init:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install go.uber.org/mock/mockgen@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: dev
dev:
	air -c .air.toml

.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(CTL_NAME) ./cmd/$(CTL_NAME)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(WORKER_NAME) ./cmd/$(WORKER_NAME)

.PHONY: build-all
build-all:
	@mkdir -p $(BIN_DIR)
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-amd64   ./cmd/$(APP_NAME)
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-arm64   ./cmd/$(APP_NAME)
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64  ./cmd/$(APP_NAME)
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64  ./cmd/$(APP_NAME)

.PHONY: test
test:
	go test -v -race -count=1 -coverprofile=coverage.out -covermode=atomic -shuffle=on ./...

.PHONY: test-integration
test-integration:
	go test -v -race -count=1 -tags=integration -coverprofile=coverage.out ./...

.PHONY: cover
cover: test
	go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml ./...

.PHONY: sec
sec:
	gosec -quiet -fmt=sarif -out=gosec.sarif ./...

.PHONY: mock
mock:
	mockgen -source=internal/queue/client.go -destination=tests/mocks/mock_queue_client.go -package=mocks
	mockgen -source=internal/domain/message/repository.go -destination=tests/mocks/mock_message_repo.go -package=mocks

.PHONY: proto
proto:
	buf generate

.PHONY: migrate-up
migrate-up:
	@test -n "$(DATABASE_URL)" || (echo "ERROR: DATABASE_URL is not set" && exit 1)
	migrate -path database/migrations -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down:
	@test -n "$(DATABASE_URL)" || (echo "ERROR: DATABASE_URL is not set" && exit 1)
	migrate -path database/migrations -database "$(DATABASE_URL)" down 1

.PHONY: migrate-create
migrate-create:
	@test -n "$(NAME)" || (echo "ERROR: NAME is required" && exit 1)
	migrate create -ext sql -dir database/migrations -seq $(NAME)

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)/ tmp/ dist/ coverage.out coverage.html gosec.sarif build-errors.log

.PHONY: docker-build
docker-build:
	docker build -t wispmail:latest -f build/Dockerfile .

.PHONY: docker-run
docker-run:
	docker compose -f build/docker-compose.yml up -d

.PHONY: docker-stop
docker-stop:
	docker compose -f build/docker-compose.yml down

.PHONY: version
version:
	@echo "Version:    $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)"
	@echo "Commit:     $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)"
	@echo "Build Time: $(shell date -u +%Y-%m-%dT%H:%M:%SZ)"
	@echo "Go Version: $(shell go version)"