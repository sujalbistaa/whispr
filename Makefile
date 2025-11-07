.PHONY: dev build

# Default target
all: build

# Run the server in development mode
dev:
	@echo "Starting server in dev mode..."
	# We use `go run` to compile and run in one step
	go run ./cmd/server

# Build the production binary
build:
	@echo "Building binary..."
	# This creates a single executable file in the 'bin' directory
	CGO_ENABLED=0 go build -o ./bin/whispr -ldflags="-w -s" ./cmd/server

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -f ./bin/whispr
	rm -f whispr.db whispr.db-journal