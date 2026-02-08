.PHONY: build run clean test

# Build the application
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o hctf2 cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go --port 8090 --admin-email admin@hctf.local --admin-password changeme

# Run with existing database
run-dev:
	go run cmd/server/main.go --port 8090

# Clean build artifacts
clean:
	rm -f hctf2 hctf2.db

# Run tests
test:
	go test ./... -v

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Create a production build
build-prod:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o hctf2-linux-amd64 cmd/server/main.go

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  run         - Run with admin setup"
	@echo "  run-dev     - Run without admin setup"
	@echo "  clean       - Remove build artifacts"
	@echo "  test        - Run tests"
	@echo "  deps        - Install dependencies"
	@echo "  fmt         - Format code"
	@echo "  build-prod  - Build for production (Linux AMD64)"
