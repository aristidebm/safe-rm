BINARY := safe-rm
BUILD_DIR := build
GO := go

.PHONY: all build test format run install clean

all: build

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY) .

test:
	$(GO) test ./...

format:
	$(GO) fmt ./...

run:
	$(GO) run . -- $(ARGS)

install:
	$(GO) install .

clean:
	rm -rf $(BUILD_DIR)
