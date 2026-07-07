BINARY      := oci-compute-capacity-report
VERSION     ?= $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "4.0.0")
BUILD_DIR   := bin
GO          := go
UPX         := upx
LDFLAGS     := -s -w -X github.com/Olygo/OCI_ComputeCapacityReport/internal/config.Version=$(VERSION)
UPX_FLAGS   := --best --lzma

.PHONY: all build build-noupx clean deps install help run

all: build

build: $(BUILD_DIR)/$(BINARY)

$(BUILD_DIR)/$(BINARY): $(shell find . -name '*.go' -not -path './vendor/*')
	@mkdir -p $(BUILD_DIR)
	@echo "==> Building $(BINARY) v$(VERSION)"
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY).tmp .
	@echo "==> Compressing with UPX"
	@rm -f $(BUILD_DIR)/$(BINARY)
	$(UPX) $(UPX_FLAGS) -o $(BUILD_DIR)/$(BINARY) $(BUILD_DIR)/$(BINARY).tmp
	@rm -f $(BUILD_DIR)/$(BINARY).tmp
	@echo "==> Done: $(BUILD_DIR)/$(BINARY)"

build-noupx:
	@mkdir -p $(BUILD_DIR)
	@echo "==> Building $(BINARY) v$(VERSION) (without UPX)"
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) .
	@echo "==> Done: $(BUILD_DIR)/$(BINARY)"

deps:
	@echo "==> Downloading Go dependencies"
	$(GO) mod download

clean:
	@echo "==> Cleaning"
	rm -rf $(BUILD_DIR)/$(BINARY) $(BUILD_DIR)/$(BINARY).tmp $(BUILD_DIR)

install: build
	@echo "==> Installing to /usr/local/bin"
	install -m 755 $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)

run: build
	./$(BUILD_DIR)/$(BINARY)

help:
	@echo "Targets:"
	@echo "  build        Build and compress binary with UPX (default)"
	@echo "  build-noupx  Build without UPX compression"
	@echo "  deps         Download Go module dependencies"
	@echo "  clean        Remove build artifacts"
	@echo "  install      Install binary to /usr/local/bin"
	@echo "  run          Build and run the application"
	@echo "  help         Show this help"