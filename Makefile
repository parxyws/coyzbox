DB_USER ?= postgres
DB_PASSWORD ?= s408RZ1ej76vl9ta
DB_HOST ?= localhost
DB_PORT ?= 5500
DB_NAME ?= cozybox
DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

MOCKERY_VERSION ?= latest

.PHONY: migrate-up migrate-down migrate-down-force swagger ui-build build mock install-tools test compose-up

compose-up:
	podman compose -f deployments/docker-compose.yaml up -d

migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down

migrate-down-force:
	migrate -path db/migrations -database "$(DB_URL)" force 1

swagger:
	swag init -g cmd/cozybox/main.go -o docs/

ui-build:
	cd web && pnpm build

build: ui-build
	go build -o cozybox ./cmd/cozybox

install-tools:
	go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)

mock:
	mockery

test:
	go test ./test/... -count=1 -v
