.PHONY: dev run build test test-cover lint docker-up docker-down docker-build clean

# Application
APP_NAME=spotgov-pipeline
MAIN_PATH=./cmd/server

dev:
	go run $(MAIN_PATH)/main.go

run: build
	./bin/$(APP_NAME)

build:
	CGO_ENABLED=0 go build -o bin/$(APP_NAME) $(MAIN_PATH)/main.go

test:
	go test -v -race -count=1 ./...

test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

# Docker
docker-build:
	docker build -t $(APP_NAME) .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-down-v:
	docker compose down -v

clean:
	rm -rf bin/ coverage.out coverage.html
