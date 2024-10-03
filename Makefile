.PHONY: build run test lint

build:
	go build -o bin/server cmd/main.go

run:
	go run cmd/main.go

test:
	go test -v ./...

lint:
	golangci-lint run

.DEFAULT_GOAL := build