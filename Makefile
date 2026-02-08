BINARY_NAME=ankies-franc
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Installation directory (can be overridden)
PREFIX?=/usr/local
BINDIR?=$(PREFIX)/bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

.PHONY: all build clean test coverage lint install install-user uninstall deps help

all: build

## build: Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

## install: Install to $(BINDIR) (default: /usr/local/bin)
install: build
	@echo "Installing $(BINARY_NAME) to $(BINDIR)..."
	@mkdir -p $(BINDIR)
	@cp $(BINARY_NAME) $(BINDIR)/$(BINARY_NAME)
	@chmod +x $(BINDIR)/$(BINARY_NAME)
	@echo "Done! Run '$(BINARY_NAME)' from anywhere."

## install-user: Install to GOBIN (no sudo required)
install-user: build
	@GOBIN=$$(go env GOBIN); \
	if [ -z "$$GOBIN" ]; then GOBIN=$$(go env GOPATH)/bin; fi; \
	echo "Installing $(BINARY_NAME) to $$GOBIN..."; \
	mkdir -p "$$GOBIN"; \
	cp $(BINARY_NAME) "$$GOBIN/$(BINARY_NAME)"; \
	chmod +x "$$GOBIN/$(BINARY_NAME)"; \
	echo "Done!"

## uninstall: Remove from $(BINDIR)
uninstall:
	@echo "Removing $(BINARY_NAME) from $(BINDIR)..."
	@rm -f $(BINDIR)/$(BINARY_NAME)
	@echo "Done!"

## clean: Remove build artifacts
clean:
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "Cleaned."

## test: Run tests
test:
	$(GOTEST) -v ./...

## coverage: Run tests with coverage report
coverage:
	$(GOTEST) ./... -coverprofile=coverage.out -covermode=atomic
	$(GOCMD) tool cover -func=coverage.out
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## lint: Run linter
lint:
	golangci-lint run ./...

## deps: Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'
