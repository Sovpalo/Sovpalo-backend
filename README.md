# sovpalo-backend

Минимальный REST API каркас на Go (Gin) c Clean Architecture и подключением к PostgreSQL.

## Быстрый старт

1) Поднять базу данных:

```bash
docker compose up -d
```

2) Задать переменные окружения (пример в `.env.example`).

3) Применить миграции:

```bash
go run ./cmd/migrate up
```

4) Запустить сервер:

```bash
go run ./cmd
```

## Миграции

Команды запускаются через `cmd/migrate` (используется goose):

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down
go run ./cmd/migrate status
```

## Эндпоинты

- `GET /health` — проверка доступности сервиса и базы данных.
