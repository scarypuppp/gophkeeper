BUILD_INFO_PKG := github.com/scarypuppp/gophkeeper/internal/buildinfo

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)

LDFLAGS := -X '$(BUILD_INFO_PKG).version=$(VERSION)' \
           -X '$(BUILD_INFO_PKG).date=$(BUILD_DATE)' \
           -X '$(BUILD_INFO_PKG).commit=$(COMMIT)'

env:
	cp .env.example .env

build: build-client build-server

build-client:
	go build -ldflags "$(LDFLAGS)" -o ./bin/gophkeeper ./cmd/client/

build-server:
	go build -o ./bin/gophkeeper-server ./cmd/server/

run-server:
	go run ./cmd/server/main.go

test:
	go test ./...

generate-swagger:
	swag init -g ./cmd/server/main.go -o docs

.PHONY: env build build-client build-server run-server test generate-swagger
