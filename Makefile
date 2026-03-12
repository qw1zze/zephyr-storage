.PHONY: run build test docker-up docker-down tidy

run:
	go run ./cmd/server

build:
	go build -o ./bin/server ./cmd/server

test:
	go test ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

tidy:
	go mod tidy
