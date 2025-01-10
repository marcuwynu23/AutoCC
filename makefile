# Makefile for building and managing the autocc executable

# Variables
APP_NAME := autocc
SRC_DIR := ./
BUILD_DIR := ./build
BIN_DIR := /usr/local/bin
CONFIG_DIR := /etc/autocc
EXAMPLES_DIR := ~/autocc/examples
SCRIPTS_DIR := ~/autocc/scripts
WORKING_DIR := ~/autocc

# Default target
.PHONY: all
all: build

# Build the executable
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go build -o $(BUILD_DIR)/$(APP_NAME) $(SRC_DIR)
	@echo "Build complete. Executable located at $(BUILD_DIR)/$(APP_NAME)."

# Install the executable for Linux
.PHONY: install
install: build
	@echo "Installing $(APP_NAME) to $(BIN_DIR)..."
	@sudo mkdir -p $(CONFIG_DIR)
	@sudo mkdir -p $(WORKING_DIR)
	@sudo cp -r ./examples $(EXAMPLES_DIR)
	@sudo cp -r ./scripts $(SCRIPTS_DIR)
	@sudo cp ./settings.json $(CONFIG_DIR)/settings.json
	@sudo cp $(BUILD_DIR)/$(APP_NAME) $(BIN_DIR)/$(APP_NAME)
	@sudo chmod +x $(BIN_DIR)/$(APP_NAME)
	@echo "$(APP_NAME) installed successfully. Working directory set to $(WORKING_DIR)."

# Install the executable for Windows
.PHONY: install-windows
install-windows: build
	@echo "Installing $(APP_NAME) for Windows..."
	@mkdir -p C:\autocc\examples
	@mkdir -p C:\autocc\scripts
	@mkdir -p C:\autocc\config
	copy .\examples C:\autocc\examples /E /Y
	copy .\scripts C:\autocc\scripts /E /Y
	copy .\settings.json C:\autocc\config\settings.json
	copy $(BUILD_DIR)\$(APP_NAME).exe C:\autocc\$(APP_NAME).exe
	@echo "$(APP_NAME) installed successfully in C:\\autocc."

# Clean up build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete."

# Uninstall the executable
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(APP_NAME) from $(BIN_DIR) and cleaning up..."
	@sudo rm -f $(BIN_DIR)/$(APP_NAME)
	@sudo rm -rf $(CONFIG_DIR)
	@sudo rm -rf $(WORKING_DIR)
	@echo "$(APP_NAME) uninstalled successfully."

# Uninstall the executable for Windows
.PHONY: uninstall-windows
uninstall-windows:
	@echo "Uninstalling $(APP_NAME) for Windows..."
	@rmdir /S /Q C:\autocc
	@echo "$(APP_NAME) uninstalled successfully."

# Run the application
.PHONY: run
run: build
	@echo "Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)
