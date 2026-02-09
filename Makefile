APP_NAME=miniapi
CMD=./cmd/server

ALLOWED_SSLMODES := disable allow prefer require verify-ca verify-full

ifneq (,$(wildcard ./.env))
include .env
export
endif

MIGRATE_VERSION=4.17.1
MIGRATE_BIN=./bin/migrate
MIGRATIONS_DIR=./migrations
DB_SSLMODE ?= disable
DB_SSLMODE_STR := $(if $(filter 1,$(DB_SSLMODE)),require,disable)
DB_URL := postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE_STR)

M = $(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)"
DC = docker compose
DCM = $(DC) run --rm --entrypoint /app/migrate api -path /app/migrations -database "$(DB_URL)"

.PHONY: help
help:
	@echo " check-env - check env variables"
	@echo "Local:"
	@echo " deps      - download go deps"
	@echo " run       - run server locally (requires .env in project root)"
	@echo " fmt       - gofmt code"
	@echo " test      - run tests"
	@echo "migrations:"
	@echo " migrate-install - install golang-migrate"
	@echo " migrate-up      - up migrates"
	@echo " migrate-down    - down 1 migrate"
	@echo " migrate-force   - force migrate"
	@echo " migrate-version - version"
	@echo ""
	@echo "Docker:"
	@echo " d-build   - docker compose build (force)"
	@echo " d-up      - start db+api via docker compose"
	@echo " d-down    - stop stack"
	@echo " d-logs    - tail api logs"
	@echo "migrations:"
	@echo " d-migrate-up      - up migrates"
	@echo " d-migrate-down    - down 1 migrate"
	@echo " d-migrate-force   - force migrate"
	@echo " d-migrate-version - version"
	@echo ""
	@echo "Tests:"
	@echo " test-unit          - run UNIT tests"
	@echo " test-integration   - run integration tests"

.PHONY: check-env
check-env:
	@if [ -z "$(DB_HOST)" ] || [ -z "$(DB_PORT)" ] || [ -z "$(DB_NAME)" ] || [ -z "$(DB_USERNAME)" ] || [ -z "$(DB_PASSWORD)" ]; then \
		echo "Missing DB_* env vars. Required: DB_HOST DB_PORT DB_NAME DB_USERNAME DB_PASSWORD"; \
		exit 1; \
	fi
	@if [ -z "$(DB_SSLMODE)" ]; then \
		echo "Missing DB_SSLMODE. Allowed: $(ALLOWED_SSLMODES)"; \
		exit 1; \
	fi
	@if ! echo "$(ALLOWED_SSLMODES)" | tr ' ' '\n' | grep -qx "$(DB_SSLMODE)"; then \
		echo "Invalid DB_SSLMODE=$(DB_SSLMODE). Allowed: $(ALLOWED_SSLMODES)"; \
		exit 1; \
	fi

.PHONY: deps run test fmt
deps:
	go mod download

run:
	go run $(CMD)

test:
	go test ./...

fmt:
	gofmt -w .

.PHONY: d-build d-up d-down d-logs
d-build: check-env
	$(DC) build --no-cache

d-up: check-env
	$(DC) up -d

d-down: check-env
	$(DC) down

d-logs: check-env
	$(DC) logs -f api

.PHONY: migrate-install migrate-up migrate-down migrate-force migrate-version
migrate-install: check-env
	@mkdir -p ./bin
	@if [ ! -f "$(MIGRATE_BIN)" ]; then \
		echo "Installing golang-migrate $(MIGRATE_VERSION)..." ; \
		GOBIN=$$(pwd)/bin go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v$(MIGRATE_VERSION); \
	else \
		echo "migrate already installed: $(MIGRATE_BIN)" ; \
	fi

migrate-up: check-env migrate-install
	$(M) up

migrate-down: check-env migrate-install
	$(M) down 1

migrate-force: check-env migrate-install
	@if [ -z "$(V)" ]; then echo "Usage: make migrate-force V=<version>"; exit 1; fi
	$(M) force $(V)

migrate-version: check-env migrate-install
	$(M) version

.PHONY: d-migrate-up d-migrate-down d-migrate-version d-migrate-force
d-migrate-up: check-env
	$(DCM) up

d-migrate-down: check-env
	$(DCM) down 1

d-migrate-version: check-env
	$(DCM) version

d-migrate-force: check-env
	@if [ -z "$(V)" ]; then echo "Usage: make d-migrate-force V=<version>"; exit 1; fi
	$(DCM) force $(V)

.PHONY: test-unit test-integration
test-unit:
	go test ./...

test-integration:
	go test -tags=integration ./...

