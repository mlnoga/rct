all: build test lint

build: *.go
	go build

test: *.go
	go test

lint: *.go
	golangci-lint run
