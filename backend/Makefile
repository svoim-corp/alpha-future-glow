.PHONY: build run test vet fmt swagger up

build:
	go build ./...

run:
	go run ./cmd/api

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l .

# Требуется swag: go install github.com/swaggo/swag/cmd/swag@latest
swagger:
	swag init -g cmd/api/main.go -o docs --parseInternal

up:
	docker compose -f ../deploy/docker-compose.yml up --build
