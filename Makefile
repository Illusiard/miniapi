APP_NAME=miniapi
CMD=./cmd/server

.PHONY: help
help:
	@echo "Targets:"
	@echo "  make deps      - download go deps"
	@echo "  make run       - run server locally (requires .env in project root)"
	@echo "  make fmt       - gofmt code"
	@echo "  make test      - run tests"
	@echo "  make d-build   - docker compose build (force)"
	@echo "  make d-up      - start db+api via docker compose"
	@echo "  make d-down    - stop stack"
	@echo "  make d-logs    - tail api logs"

.PHONY: deps
deps:
	go mod download

.PHONY: run
run:
	go run $(CMD)

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: d-build
d-build:
	docker compose build --no-cache

.PHONY: d-up
d-up:
	docker compose up -d

.PHONY: d-down
d-down:
	docker compose down

.PHONY: d-logs
d-logs:
	docker compose logs -f api
