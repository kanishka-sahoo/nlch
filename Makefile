# Makefile for cross-platform build and install of nlch

APP_NAME = nlch

PLATFORMS = \
	"linux/amd64" \
	"linux/arm64" \
	"windows/amd64" \
	"darwin/amd64" \
	"darwin/arm64"

BIN_DIR = bin

.PHONY: all build clean install

all: build

build:
	@mkdir -p $(BIN_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		output_name=$(BIN_DIR)/$(APP_NAME)-$${platform%/*}-$${platform#*/}; \
		if [ "$${platform%/*}" = "windows" ]; then output_name=$$output_name.exe; fi; \
		echo "Building $$output_name"; \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} go build -o $$output_name . ; \
	done

clean:
	rm -rf $(BIN_DIR)

install:
	@echo "Installing $(APP_NAME) binary for your OS..."
	@GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) \
	output_name=$(BIN_DIR)/$(APP_NAME)-$(shell go env GOOS)-$(shell go env GOARCH); \
	if [ "$(shell go env GOOS)" = "windows" ]; then output_name=$$output_name.exe; fi; \
	if [ ! -f $$output_name ]; then \
		echo "Binary not found for your OS/ARCH. Run 'make build' first."; \
		exit 1; \
	fi; \
	if [ "$(shell go env GOOS)" = "windows" ]; then \
		install_dir="$$USERPROFILE\\bin"; \
		mkdir -p "$$install_dir"; \
		cp "$$output_name" "$$install_dir\\$(APP_NAME).exe"; \
		echo "Installed to $$install_dir\\$(APP_NAME).exe"; \
	else \
		install_dir="/usr/local/bin"; \
		cp "$$output_name" "$$install_dir/$(APP_NAME)"; \
		chmod +x "$$install_dir/$(APP_NAME)"; \
		echo "Installed to $$install_dir/$(APP_NAME)"; \
	fi
