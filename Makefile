# jumblejuice makefile
# automated build, test, and release tasks

.PHONY: all build test clean install setup coverage bench lint release help

# variables
BINARY_NAME := jj
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GOFILES := $(shell find . -name "*.go" -type f)
PACKAGES := $(shell go list ./...)
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # no color

# default target
all: clean build test

## help: show this help message
help:
	@echo "JumbleJuice Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /' | column -t -s ':'

## setup: install dependencies and tools
setup:
	@echo "$(GREEN)Setting up development environment...$(NC)"
	go mod download
	go install github.com/spf13/cobra@latest
	go install github.com/spf13/viper@latest
	go install github.com/stretchr/testify@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)Setup complete!$(NC)"

## build: build the binary
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	go build $(LDFLAGS) -o $(BINARY_NAME) cmd/jj/main.go
	@echo "$(GREEN)Build complete: ./$(BINARY_NAME)$(NC)"

## test: run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	@if command -v go &> /dev/null; then \
		go test -v -race -timeout 30s $(PACKAGES); \
	else \
		echo "$(YELLOW)Go not installed - running simulation tests$(NC)"; \
		./run_tests.sh; \
	fi
	@echo "$(GREEN)Tests complete!$(NC)"

## test-short: run short tests only
test-short:
	@echo "$(GREEN)Running short tests...$(NC)"
	go test -v -short -timeout 10s $(PACKAGES)

## coverage: run tests with coverage
coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	go test -v -race -coverprofile=coverage.out -covermode=atomic $(PACKAGES)
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"
	@echo "Total coverage: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"

## bench: run benchmarks
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	go test -bench=. -benchmem -timeout 10m $(PACKAGES)

## fuzz: run fuzz tests
fuzz:
	@echo "$(GREEN)Running fuzz tests...$(NC)"
	go test -fuzz=FuzzEncoders -fuzztime=30s ./internal/encoder
	go test -fuzz=FuzzEmitters -fuzztime=30s ./internal/emitter

## lint: run linters
lint:
	@echo "$(GREEN)Running linters...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not installed, running go vet...$(NC)"; \
		go vet $(PACKAGES); \
	fi
	@if command -v staticcheck > /dev/null; then \
		staticcheck $(PACKAGES); \
	else \
		echo "$(YELLOW)staticcheck not installed, skipping...$(NC)"; \
	fi
	@echo "$(GREEN)Linting complete!$(NC)"

## fmt: format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt $(PACKAGES)
	gofmt -s -w $(GOFILES)
	@echo "$(GREEN)Formatting complete!$(NC)"

## clean: clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning...$(NC)"
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf dist/
	go clean -cache
	@echo "$(GREEN)Clean complete!$(NC)"

## install: install binary to /usr/local/bin
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME) to /usr/local/bin...$(NC)"
	@if [ -w /usr/local/bin ]; then \
		cp $(BINARY_NAME) /usr/local/bin/; \
	else \
		echo "$(YELLOW)Need sudo access to install to /usr/local/bin$(NC)"; \
		sudo cp $(BINARY_NAME) /usr/local/bin/; \
	fi
	@echo "$(GREEN)Installation complete!$(NC)"

## uninstall: remove binary from /usr/local/bin
uninstall:
	@echo "$(YELLOW)Uninstalling $(BINARY_NAME)...$(NC)"
	@if [ -w /usr/local/bin/$(BINARY_NAME) ]; then \
		rm -f /usr/local/bin/$(BINARY_NAME); \
	else \
		sudo rm -f /usr/local/bin/$(BINARY_NAME); \
	fi
	@echo "$(GREEN)Uninstall complete!$(NC)"

## dev: run in development mode with hot reload
dev:
	@echo "$(GREEN)Starting development mode...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)Air not installed. Install with: go install github.com/cosmtrek/air@latest$(NC)"; \
		go run cmd/jj/main.go; \
	fi

## run: run the application
run: build
	./$(BINARY_NAME)

## docker-build: build docker image
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t jumblejuice:$(VERSION) .
	@echo "$(GREEN)Docker build complete!$(NC)"

## docker-run: run in docker container
docker-run: docker-build
	docker run --rm -it jumblejuice:$(VERSION)

## release: create release binaries
release:
	@echo "$(GREEN)Building release binaries...$(NC)"
	@mkdir -p dist

	# linux amd64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 cmd/jj/main.go

	# linux arm64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 cmd/jj/main.go

	# macos amd64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 cmd/jj/main.go

	# macos arm64 (m1/m2)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 cmd/jj/main.go

	# windows amd64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe cmd/jj/main.go

	@echo "$(GREEN)Release binaries created in dist/$(NC)"
	@ls -lh dist/

## checksum: generate checksums for release binaries
checksum: release
	@echo "$(GREEN)Generating checksums...$(NC)"
	cd dist && sha256sum * > checksums.txt
	@cat dist/checksums.txt

## ci: run ci pipeline locally
ci: clean lint test coverage
	@echo "$(GREEN)CI pipeline complete!$(NC)"

## check: quick check before commit
check: fmt lint test-short
	@echo "$(GREEN)Pre-commit checks passed!$(NC)"

## deps: show dependencies
deps:
	@echo "$(GREEN)Project dependencies:$(NC)"
	go mod graph

## update: update dependencies
update:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	go get -u ./...
	go mod tidy
	@echo "$(GREEN)Dependencies updated!$(NC)"

## version: show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"

## stats: show code statistics
stats:
	@echo "$(GREEN)Code statistics:$(NC)"
	@echo "Lines of Go code: $$(find . -name '*.go' -type f | xargs wc -l | tail -1 | awk '{print $$1}')"
	@echo "Number of Go files: $$(find . -name '*.go' -type f | wc -l)"
	@echo "Number of packages: $$(go list ./... | wc -l)"
	@echo "Number of tests: $$(go test ./... -list . | grep -c '^Test')"

## todo: show todo items in code
todo:
	@echo "$(GREEN)TODO items:$(NC)"
	@grep -rn "TODO\|FIXME\|XXX" --include="*.go" . || echo "No TODOs found!"

## examples: build example files
examples:
	@echo "$(GREEN)Building examples...$(NC)"
	@mkdir -p examples/output
	@if [ -f $(BINARY_NAME) ]; then \
		echo "test data" > examples/test.bin; \
		./$(BINARY_NAME) emit --in examples/test.bin --encoder hex > examples/output/hex.go; \
		./$(BINARY_NAME) emit --in examples/test.bin --encoder b64 > examples/output/b64.go; \
		./$(BINARY_NAME) emit --in examples/test.bin --encoder dec > examples/output/dec.go; \
		echo "$(GREEN)Examples created in examples/output/$(NC)"; \
	else \
		echo "$(RED)Build the binary first with 'make build'$(NC)"; \
	fi

## watch: watch for changes and rebuild
watch:
	@echo "$(GREEN)Watching for changes...$(NC)"
	@while true; do \
		make build test-short; \
		echo "$(YELLOW)Waiting for changes...$(NC)"; \
		fswatch -1 -r --exclude .git --exclude $(BINARY_NAME) .; \
	done

# hidden targets for development
.PHONY: debug
debug: LDFLAGS += -gcflags="all=-N -l"
debug: build
	@echo "$(GREEN)Debug build complete!$(NC)"

.PHONY: prof
prof:
	@echo "$(GREEN)Running with profiling...$(NC)"
	go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=.
	@echo "View CPU profile: go tool pprof cpu.prof"
	@echo "View memory profile: go tool pprof mem.prof"
