# Build worker and enqueue
build:
	docker build -t go_work_horse_worker .
	go build -o bin/enqueue ./cmd/enqueue/main.go

# Start all services (worker, redis, prometheus, grafana, jaeger)
up:
	docker-compose up --build

down:
	docker-compose down

# Unit tests
unit-test:
	go test ./pkg/jobqueue/...

# Integration tests
integration-test:
	go test ./test/integration/...

# All tests
all-tests:
	go test ./...

# Test coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint (using golangci-lint)
lint:
	golangci-lint run ./...

# Security (using gosec)
security:
	gosec ./...

# Install tools dependencies (golangci-lint, gosec)
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run enqueue (example usage)
enqueue:
	REDIS_ADDR=localhost:6379 ./bin/enqueue '{"foo":"bar"}'

# Target for CI: build, lint, security, tests
ci: build lint security all-tests

.PHONY: build up down unit-test integration-test all-tests coverage lint security tools enqueue ci 