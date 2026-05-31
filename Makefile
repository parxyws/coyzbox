DB_USER ?= postgres
DB_PASSWORD ?= password
DB_HOST ?= localhost
DB_PORT ?= 5500
DB_NAME ?= cozybox
DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down

migrate-down-force:
	migrate -path db/migrations -database "$(DB_URL)" force 1

.PHONY: migrate-up migrate-down migrate-down-force swagger ui-build build

swagger:
	swag init -g cmd/main.go -o docs/

ui-build:
	cd web && pnpm build

build: ui-build
	go build -o cozybox ./cmd/main.go
