.PHONY: build server tui test lint migrate-up migrate-down clean dev dev-tui test-cover migrate-create release

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS  = -X github.com/Akram012388/niotebook-tui/internal/build.Version=$(VERSION) \
           -X github.com/Akram012388/niotebook-tui/internal/build.CommitSHA=$(COMMIT)

build: server tui

server:
	go build -ldflags "$(LDFLAGS)" -o bin/niotebook-server ./cmd/server

tui:
	go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui ./cmd/tui

dev:
	go run ./cmd/server

dev-tui:
	go run ./cmd/tui --server http://localhost:8080

test:
	go test ./... -v -race

test-cover:
	go test ./... -v -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

migrate-up:
	migrate -path migrations -database "$(NIOTEBOOK_DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(NIOTEBOOK_DB_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

clean:
	rm -rf bin/ coverage.out coverage.html

release:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-server-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-server-linux-arm64 ./cmd/server
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui-darwin-amd64 ./cmd/tui
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui-darwin-arm64 ./cmd/tui
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui-linux-amd64 ./cmd/tui
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui-linux-arm64 ./cmd/tui
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui-windows-amd64.exe ./cmd/tui
