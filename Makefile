# Project directories
DIRS := check-slots-consistency hash-slot-keys migrate-slots
# Binaries will be placed in the bin folder
BIN_DIR := bin

# Build binaries for the current platform
.PHONY: build
build:
	@for dir in $(DIRS); do \
		echo "Building $$dir"; \
		go build -o $(BIN_DIR)/$$dir ./$$dir; \
	done

# Build binaries for Linux platform
.PHONY: linux-build
linux-build:
	@for dir in $(DIRS); do \
		echo "Building $$dir for Linux"; \
		GOARCH=amd64 GOOS=linux go build -o $(BIN_DIR)/linux-$$dir ./$$dir; \
	done

# Clean up the bin directory
.PHONY: clean
clean:
	@echo "Cleaning up"
	@rm -rf $(BIN_DIR)/*
