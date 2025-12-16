.PHONY: all build run clean install test help

# Variables
BINARY_NAME=subspace
GO=go
GOFLAGS=-v

all: build

## help: Display this help message
help:
	@echo "SubSpace Makefile Commands:"
	@echo ""
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application with search action"
	@echo "  make install        - Install dependencies"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make test           - Run tests (when implemented)"
	@echo "  make help           - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run ARGS='-action=search -keywords=\"Software Engineer\"'"
	@echo ""

## install: Install Go dependencies
install:
	@echo "📦 Installing dependencies..."
	$(GO) mod download
	@echo "✅ Dependencies installed"

## build: Build the application
build: install
	@echo "🔨 Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) main.go
	@echo "✅ Build complete: ./$(BINARY_NAME)"

## run: Run the application (use ARGS for custom arguments)
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	./$(BINARY_NAME) $(ARGS)

## clean: Remove build artifacts and temporary files
clean:
	@echo "🧹 Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf data/*.db
	@rm -rf data/sessions/*
	@echo "✅ Clean complete"

## test: Run tests
test:
	@echo "🧪 Running tests..."
	$(GO) test -v ./...

## setup: Run the setup script
setup:
	@echo "⚙️  Running setup..."
	./setup.sh

## search: Quick search command
search: build
	./$(BINARY_NAME) -action=search -keywords="$(KEYWORDS)" -max-results=$(MAX_RESULTS)

## connect: Quick connect command
connect: build
	./$(BINARY_NAME) -action=connect -keywords="$(KEYWORDS)" -max-results=$(MAX_RESULTS) -message="$(MESSAGE)"

## message: Quick message command
message: build
	./$(BINARY_NAME) -action=message -message="$(MESSAGE)"

# Default values for quick commands
KEYWORDS ?= "Software Engineer"
MAX_RESULTS ?= 10
MESSAGE ?= "Hi {firstName}, I'd love to connect!"
