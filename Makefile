GOCMD=go
GOBUILD=$(GOCMD) build
GOMOD=$(GOCMD) mod

# Binary
BIN_FOLDER=$(shell pwd)/bin
BIN_NAME=$(shell basename "$(PWD)")

pulsar: deps
	@echo "Building $(BIN_NAME)..."
	@$(GOBUILD) -o $(BIN_FOLDER)/$(BIN_NAME) src/main.go
	@echo "Done"

deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@echo "Done"

clean:
	@$(GOCMD) clean

.PHONY: pulsar deps clean
