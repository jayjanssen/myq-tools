.PHONY: install test test-race test-verbose test-coverage benchmark benchmark-ci build clean

# Install all myq-* binaries
install:
	@for dir in $(shell ls -d myq-*/); do \
		cd $$dir && go install .; \
		cd ..; \
	done

# Run all tests
test:
	go test ./...

# Run tests with race detector
test-race:
	go test -race ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage report
test-coverage:
	go test -cover ./...

# Run benchmarks
benchmark:
	go test -bench=. -benchmem ./...

# Run benchmarks (fast CI smoke test)
benchmark-ci:
	go test -bench=. -benchtime=1x ./...

# Build all binaries
build:
	@for dir in $(shell ls -d myq-*/); do \
		echo "Building $$dir..."; \
		cd $$dir && go build -o ../bin/$$(basename $$dir) .; \
		cd ..; \
	done

# Clean build artifacts
clean:
	rm -rf bin/
	go clean ./...
