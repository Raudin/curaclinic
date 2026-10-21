.PHONY: fmt vet lint test test-integration sqlc build run migrate seed

fmt:
	gofmt -l -w ./cmd ./internal ./migrations

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test -race -count=1 -shuffle=on ./...

test-integration:
	go test -race -count=1 ./... # requires TEST_DATABASE_URL

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate

build:
	go build ./...

run:
	go run ./cmd/api

migrate:
	go run ./cmd/migrate up

seed:
	go run ./cmd/seed --demo
