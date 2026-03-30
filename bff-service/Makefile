.PHONY: build run clean docker-build docker-up docker-down

APP_NAME=bff-service
BUILD_DIR=.

build:
	go build -o $(APP_NAME) ./server/

run: build
	./$(APP_NAME) grpc-gw-server

clean:
	rm -f $(APP_NAME)

unit-test:
	go test -v -coverprofile=coverage/coverage.out ./server/...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html

docker-build:
	docker build -t $(APP_NAME) .

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v
