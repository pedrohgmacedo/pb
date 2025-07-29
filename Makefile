# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOLINT=staticcheck

GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS)

# Build directory
BUILD_DIR=build

PROGRAMNAME=pb


# Default target
all: generate-version build

# Build for host OS
build: generate-version
	@echo "Building for host OS..."
	@mkdir -p $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	$(GOBUILD) -o $(BUILD_DIR)/$(GOOS)-$(GOARCH)/$(PROGRAMNAME)

# Build for Linux
build-linux: generate-version
	@echo "Building for linux/amd64..."
	@mkdir -p $(BUILD_DIR)/linux-amd64
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/linux-amd64/$(PROGRAMNAME)

# Build for Windows
build-windows: generate-version
	@echo "Building for windows/amd64..."
	@mkdir -p $(BUILD_DIR)/windows-amd64
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/windows-amd64/$(PROGRAMNAME).exe

# Build for Android
build-android: generate-version
	@echo "Building for android/arm64 (Termux)..."
	@mkdir -p $(BUILD_DIR)/android-arm64
	GOOS=android GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/android-arm64/$(PROGRAMNAME)

generate-version:
	@echo 'package util' >util/version.go
	@printf 'const GitHead = "%s"' "$(shell git rev-parse HEAD)" >>util/version.go

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

lint:
	@echo "Linting..."
	$(GOLINT) ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all/build                Build for host OS"
	@echo "  build-linux              Build for linux/amd64"
	@echo "  build-windows            Build for windows/amd64"
	@echo "  build-android            Build for android/arm64 (Termux)"
	@echo "  install                  Print install cmd"
	@echo "  test                     Run tests"
	@echo "  lint                     Run linter"
	@echo "  clean                    Clean build artifacts"
	@echo "  help                     Show this help message"

install:
	@echo mv $(BUILD_DIR)/$(GOOS)-$(GOARCH)/$(PROGRANAME) 

.PHONY: all build build-linux build-windows build-android test lint clean help
