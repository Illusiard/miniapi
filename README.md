# miniapi

Небольшой мини-API-сервер для DB-CRUD:
- HTTP liveness/readiness endpoints
- PostgreSQL connection (pgxpool)
- capability-based module system (modules receive only what they need)
- docker-compose

## Requirements
- Go (1.25+ recommended)
- Docker + Docker Compose

## Quick start (Docker)
1. Create `.env` from `example.env`:
   ```bash
   cp example.env .env
   ```

2. Start stack:

   ```bash
   make d-up
   ```

3. Check endpoints:

   ```bash
   curl -i http://localhost:8080/health
   curl -i http://localhost:8080/ready
   curl -i http://localhost:8080/ping
   curl -i http://localhost:8080/meta/entities
   ```

4. Stop stack:

   ```bash
   make d-down
   ```

## Quick start (local run + Docker DB) (ну если очень надо)

1. Start only database:

   ```bash
   docker compose up -d db
   ```

2. Use local DB host in `.env`:

   * set `DB_HOST=localhost`
   * keep `EXTERNAL_DB_PORT=5432`

3. Run server:

   ```bash
   make run
   ```

## Configuration

Вся конфигурация в .env 

### Server

* `EXTERNAL_API_PORT` (default `8080`)

### Database

* `DB_HOST` (default `db`)
* `DB_PORT` (default `5432`)
* `DB_NAME` (default `miniapi`)
* `DB_USERNAME` (default `miniapi`)
* `DB_PASSWORD` (default `miniapi`)

Это позволяет подключаться к внешней базе данных

## Modules

Модули компилируются (Go packages), но загружаются через capability-based setup.
Модуль получает только необходимое:
- Routes (регистрация HTTP запросов)
- Meta (публикация мета-данных сущностей/модулей)
- Store (опционально DB access)

### Built-in modules

- **ping** (`v0.1.0`)
  - Description: Демо модуль: endpoint `/ping` + мета-данные Ping.
  - Endpoints:
    - `GET /ping` (or `/ping/`) — возвращает `pong` с timestamp.
  - Publishes entity:
    - `Ping` (демо сущность, без таблицы и данных)

- **notes** (`v0.1.0`)
  - Description: Example CRUDL модуль с хранением в PostgreSQL (`notes` table).
  - Endpoints:
    - `GET /notes` — list notes (max 100)
    - `GET /notes/{id}` — get note
    - `POST /notes` — create note `{ "title": "...", "content": "..." }`
    - `PUT /notes/{id}` — update note `{ "title": "...", "content": "..." }`
    - `DELETE /notes/{id}` — delete note
  - Publishes entity:
    - `Note` (table `notes`)


## Endpoints

* `GET /health` — жив ли серверв вообще
* `GET /ready` — готов ли (проверка БД)
* `GET /meta/entities` — список сущностей и их описание
* `GET /ping` — просто модуль для пинга

## Project layout

* `cmd/server` — application entrypoint
* `internal` — app internals

  * `app` — app lifecycle (start/stop)
  * `config` — configuration loading
  * `db` — pgxpool connection
  * `httpserver` — HTTP server & base routes
  * `caps` — capability interfaces (Routes/Meta/Store)
  * `modules` — module contract + specs
  * `meta` — meta registry for entities
  * `store` — store implementation (pgxpool adapter)
* `modules/*` — built-in modules (compiled-in)

## License

Смотрите файл с лицензией
