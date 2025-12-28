
# --- Database ENV (ใช้ค่า default + override ได้จาก export หรือ .env)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= mydatabase
DB_USER ?= myuser
DB_PASSWORD ?= mypassword
DB_SSLMODE ?= disable

# --- ประกอบ DSN
DB_DSN = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)


# --- Migration
.PHONY: migrate-up
migrate-up:
	migrate -path migrations -database "$(DB_DSN)" -verbose up

.PHONY: migrate-down
migrate-down:
	migrate -path migrations -database "$(DB_DSN)" -verbose down 1

.PHONY: migrate-force
migrate-force:
	@read -p "Version? " v; \
	migrate -path migrations -database "$(DB_DSN)" force $$v

.PHONY: migrate-drop
migrate-drop:
	migrate -path migrations -database "$(DB_DSN)" -verbose drop


# --- Run Application
.PHONY: run
run:
	go run ./cmd/main.go

# --- Test
.PHONY: test
test:
	go test ./... -v