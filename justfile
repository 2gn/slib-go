# Default recipe to list all available recipes
default:
    @just --list

# Run the application (TUI mode by default)
run:
    go run main.go

# Build the application
build:
    go build -o slib-go main.go

# Run all tests
test:
    go test -v ./...

# Format the code
fmt:
    go fmt ./...

# Run the linter
lint:
    golangci-lint run

# Clean up build artifacts
clean:
    rm -f slib-go

# Tidy up Go modules
tidy:
    go mod tidy
