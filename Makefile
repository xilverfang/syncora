# Build the CLI
build: build-syncora


build-syncora:
	go build -o bin/syncora ./cmd/bridge

# Execute the CLI
run:
	./bin/syncora

# Run tests
test:
	go test ./cmd/bridge/... ./internal/core/...

# Clean Binaries
clean:
	rm -rf bin/*

# Install dependencies
deps:
	go mod tidy -v-