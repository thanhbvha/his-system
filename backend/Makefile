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
	
build-ws:
	go build -o bin/websocket cmd/websocket/main.go

build-worker:
	go build -o bin/worker cmd/worker/main.go

dev-ws:
	go run cmd/websocket/main.go

dev-worker:
	go run cmd/worker/main.go
