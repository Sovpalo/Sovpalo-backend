# sovpalo-backend

Минимальный REST API каркас на Go (Gin) c Clean Architecture и подключением к PostgreSQL.

## Быстрый старт

1) Поднять сервисы (DB, Redis, миграции, API):

```bash
docker compose up -d
```

2) Приложение будет доступно на `http://localhost:8000`.

## Миграции

Команды запускаются через `cmd/migrate` (используется goose):

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down
go run ./cmd/migrate status
```

## Эндпоинты

- `GET /health` — проверка доступности сервиса и базы данных.
