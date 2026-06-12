.PHONY: dev migrate test lint swag build

dev:
	air

migrate:
	go run cmd/migrate/main.go

test:
	go test ./...

lint:
	golangci-lint run

swag:
	swag init -g cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go
