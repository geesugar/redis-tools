# Binaries will be placed in the bin folder
BIN_DIR := bin

# Build binaries for the current platform
.PHONY: build
build:
	go build -o $(BIN_DIR)/redis-tools .

# Build binaries for Linux platform
.PHONY: linux-build
linux-build:
	GOARCH=amd64 GOOS=linux go build -o $(BIN_DIR)/linux-redis-tools .

# Clean up the bin directory
.PHONY: clean
clean:
	@echo "Cleaning up"
	@rm -rf $(BIN_DIR)/*
