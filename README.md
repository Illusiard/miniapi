# miniapi

Небольшой мини-API-сервер для DB-CRUD:
- HTTP liveness/readiness endpoints
- PostgreSQL connection (pgxpool)
- capability-based module system (modules receive only what they need)
- docker-compose

Назачение проекта - использование при прототипировании и разработке. 

##WARNING
Доступ к сущностям осуществляется без каких-либо авторизаций, т.к. это просто api-оболочка для разработки.

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
* `HTTP_ADDR` (default `:8080`) — адрес, на котором слушает приложение (например `:8080` или `0.0.0.0:8080`)
* `LOG_LEVEL` (default `info`)

### Database

* `DB_HOST` (default `db`)
* `DB_PORT` (default `5432`)
* `DB_NAME` (default `miniapi`)
* `DB_USERNAME` (default `miniapi`)
* `DB_PASSWORD` (default `miniapi`)
* `DB_SSLMODE` (default `disable`)

### Migrations

* `AUTO_MIGRATE` (default `0`)
* `MIGRATIONS_PATH` (default `./migrations`, в Docker по умолчанию `/app/migrations`)

Это позволяет подключаться к внешней базе данных

## Modules

### Architecture overview

Ключевая идея — capability-based архитектура. В `internal/app` создаётся `caps.Setup`, и модуль получает только нужные возможности:
- `Routes`: регистрация HTTP обработчиков (через `chi`)
- `Meta`: публикация метаданных (сущности/модули)
- `Store`: доступ к БД (опционально)
- `Log`: логгер

Мета-реестр (`internal/meta`) хранит:
- список сущностей (имя, таблица, поля, модуль)
- список модулей (имя, версия, нужен ли Store)

Данные доступны через:
- `GET /meta/entities`
- `GET /meta/modules`

Модули объявляются в `internal/app` как `modules.Spec`. Это позволяет явно описывать версию, описание и необходимость `Store`.

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
* `GET /meta/modules` — список модулей и их описание
* `GET /ping` — просто модуль для пинга

## Database & migrations

Используется PostgreSQL через `pgxpool`. Подключение собирается из `DB_*` переменных (`DB_SSLMODE` по умолчанию `disable`, а `make`-таргеты требуют его явно).

Миграции:
- находятся в `./migrations`
- применяются либо через `make migrate-*`, либо автоматически при старте сервера, если `AUTO_MIGRATE=1`
- в Docker-образе есть `migrate` бинарник и путь `MIGRATIONS_PATH` по умолчанию указывает на `/app/migrations`

## Testing

Есть два уровня:
- Unit-тесты (`go test ./...`) — базовые проверки конфигурации и мета-реестра.
- Integration-тесты (`go test -tags=integration ./...`) — поднимают PostgreSQL через Testcontainers, прогоняют миграции и тестируют CRUDL для `notes`.

## Local vs Docker workflows

Локально:
- `make run` — запускает приложение, требует `.env` (см. `example.env`)
- можно поднять только БД через `docker compose up -d db`, и указать `DB_HOST=localhost`

Docker:
- `make d-up` — запускает `db` и `api` в docker-compose
- порты маппятся через `EXTERNAL_DB_PORT` и `EXTERNAL_API_PORT`

## Design decisions & limitations

Большинство решений принято потому, что это мини-тулза для разработки/прототипирования. 

- Модули подключаются статически (на этапе сборки), никакой динамической загрузки.
- Capability-based setup уменьшает связанность и ограничивает доступ модулей к инфраструктуре.
- Авто-миграции выключены по умолчанию (`AUTO_MIGRATE=0`) — это безопаснее.
- Валидация входных данных минимальная.
- Нет авторизации/аутентификации и нет публичного API-версирования.

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
