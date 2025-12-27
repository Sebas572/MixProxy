# Makefile for cross-compiling mixproxy

.PHONY: all windows linux clean

BIN_DIR := bin

all: windows linux

windows:
	@echo "Building for Windows..."
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/mixproxy.exe -ldflags="-s -w" main.go

linux:
	@echo "Building for Linux..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/mixproxy-linux-amd64 -ldflags="-s -w" main.go

clean:
	@echo "Cleaning bin directory..."
	@rm -rf $(BIN_DIR)

help:
	@echo "Available targets:"
	@echo "  all      - Build for both Windows and Linux"
	@echo "  windows  - Build for Windows (mixproxy.exe)"
	@echo "  linux    - Build for Linux (mixproxy)"
	@echo "  clean    - Remove bin directory"
	@echo "  help     - Show this help"