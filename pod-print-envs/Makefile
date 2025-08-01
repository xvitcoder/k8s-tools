# Name of the plugin (final executable will be kubectl-$(PLUGIN_NAME))
PLUGIN_NAME := pod-print-envs
BIN_NAME := kubectl-$(PLUGIN_NAME)
BIN_DIR := $(HOME)/.local/bin
SRC := main.go

.PHONY: all build install clean uninstall

# Default target: build and install
all: build install

# Build the binary
build:
	go build -o $(BIN_NAME) $(SRC)

# Install into ~/.local/bin
install: build
	mkdir -p $(BIN_DIR)
	mv -f $(BIN_NAME) $(BIN_DIR)/
	chmod +x $(BIN_DIR)/$(BIN_NAME)
	@echo "Installed as kubectl plugin: $(BIN_DIR)/$(BIN_NAME)"
	@echo "You can now run it as: kubectl $(PLUGIN_NAME)"

# Remove built binary
clean:
	rm -f $(BIN_NAME)

# Uninst

