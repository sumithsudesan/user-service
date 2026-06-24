SERVICE = notification-service
IMAGE   = $(SERVICE)
TAG     = latest

.PHONY: build test run docker-build docker-run clean

build:
	go build ./...

test:
	go test ./...

run:
	go run ./cmd

docker-build:
	docker build \
		-t $(IMAGE):$(TAG) \
		-f docker/Dockerfile \
		..

docker-run:
	docker run --rm \
		--name $(SERVICE) \
		$(IMAGE):$(TAG)

clean:
	go clean ./...
