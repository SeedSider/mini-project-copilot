DOCKERCMD=docker

DOCKER_CONTAINER_NAME?=saving-service
DOCKER_CONTAINER_IMAGE?=saving-service:latest
DOCKER_BUILD_ARGS?=

BUILD_DATE?=$(shell date -u +'%Y-%m-%dT00:00:00Z')
BUILD_VERSION?=0.1.0

BINARY=saving-service

.PHONY: build vet unit-test run depend docker-build clean

build:
	@echo "Executing go build"
	go build -v -buildmode=pie -ldflags "-X main.version=$(BUILD_VERSION)" -o app ./server/
	@echo "Binary ready"

vet:
	@echo "Running Go static code analysis with go vet"
	go vet ./...
	@echo "go vet complete"

unit-test:
	@echo "Executing go unit test"
	go test -v -json -count=1 -parallel=4 ./...
	@echo "Unit test done"

run:
	go run ./server/ grpc-gw-server --port1 9303 --port2 8081

depend:
	@echo "Pulling all Go dependencies"
	go mod download
	go mod verify
	go mod tidy
	@echo "You can now run 'make build' to compile all packages"

docker-build:
	$(DOCKERCMD) build -t $(DOCKER_CONTAINER_IMAGE) --build-arg BUILD_VERSION=$(BUILD_VERSION) $(DOCKER_BUILD_ARGS) .

clean:
	rm -f $(BINARY)
	rm -f $(BINARY).exe
	rm -f app
