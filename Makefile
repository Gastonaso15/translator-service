.PHONY: all build test coverage clean run lint

BINARY_NAME ?= translator-service
BIN_DIR ?= bin

all: test build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/translator-service

test:
	go test -v -race ./...

coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf $(BIN_DIR) coverage.out coverage.html $(BINARY_NAME)

run:
	go run ./cmd/translator-service

lint:
	golangci-lint run ./...
