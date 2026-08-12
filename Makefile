.PHONY: run build test clean fmt vet

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

test:
	go test -v -race ./...

fmt:
	go fmt ./...

vet:
	go vet ./...