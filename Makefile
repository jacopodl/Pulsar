GOCMD=go
GOBUILD=$(GOCMD) build
GOGET=$(GOCMD) get

# Binary
BIN_FOLDER=$(shell pwd)/bin
BIN_NAME=$(shell basename "$(PWD)")

export GOPATH=$(shell pwd)

pulsar: deps
	@echo "Building $(BIN_NAME)..."
	@$(GOBUILD) -o $(BIN_FOLDER)/$(BIN_NAME) src/main.go
	@echo "Done"

deps:
	@echo "Downloading dependencies..."
	@$(GOGET) golang.org/x/net/icmp
	@echo "Done"

clean:
	@$(GOCMD) clean

.PHONY: pulsar deps clean
