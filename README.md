# sovpalo-backend

Минимальный REST API каркас на Go (Gin) c Clean Architecture и подключением к PostgreSQL.

## Быстрый старт

1) Создать `.env` на основе примера и при необходимости скорректировать значения:

```bash
cp .env.example .env
```

Если вы уже поднимали PostgreSQL через Docker с другими учетными данными, сохраненный `db-data` volume продолжит использовать старый пароль. В таком случае либо выставьте в `.env` те же `DB_USER`/`DB_PASSWORD`, что использовались при первом запуске, либо пересоздайте volume.

1) Поднять сервисы (DB, Redis, миграции, API):

```bash
docker compose up -d
```

2) Приложение будет доступно на `http://localhost:8000`.

Для регистрации, входа и восстановления пароля через email перед запуском API нужно задать SMTP-переменные:

```bash
SMTP_HOST=smtp.mail.ru
SMTP_PORT=465
SMTP_USERNAME=sovpalodevteam@mail.ru
SMTP_PASSWORD=app_password
SMTP_FROM=sovpalodevteam@mail.ru
SMTP_SSL=true
SMTP_FORCE_IPV4=true
SMTP_TIMEOUT_SEC=20
SMTP_SKIP_TLS_VERIFY=false
JWT_SECRET=change_me
PASSWORD_SALT=change_me
```

Для локальной базы должны совпадать переменные приложения и контейнера PostgreSQL:

```bash
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=sovpalo
DB_HOST=localhost
DB_PORT=5433
DB_SSLMODE=disable
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
- `GET /health/smtp` — проверка SMTP-подключения и SMTP-аутентификации.
- `POST /auth/sign-up` — начало регистрации. Принимает `username`, `email`, `password`, отправляет 4-значный код на email.
- `POST /auth/sign-up/verify` — подтверждение кода. Принимает `email`, `code`, создаёт пользователя и возвращает JWT.
- `POST /auth/sign-up/resend` — повторная отправка нового 4-значного кода на email.
- `POST /auth/sign-in` — вход по `email` и `password`, сразу возвращает JWT.
- `POST /auth/password/forgot` — запуск восстановления пароля по `email`, отправляет 4-значный код на email.
- `POST /auth/password/verify` — подтверждение кода и установка нового пароля. Принимает `email`, `code`, `new_password`.
- `POST /auth/password/resend` — повторная отправка кода для восстановления пароля.
- `GET /auth/me` — получение информации о текущем пользователе. Требует `Authorization: Bearer <jwt>`, возвращает `email`, `username` и `avatar_url`.
- `POST /auth/me/avatar` — загрузка аватарки текущего пользователя. Требует `Authorization: Bearer <jwt>` и `multipart/form-data` с полем `avatar`. Поддерживаются PNG/JPEG/WEBP/GIF до 5 MB.
- `DELETE /auth/me/avatar` — удаление аватарки текущего пользователя. Требует `Authorization: Bearer <jwt>`.
- `DELETE /auth/me` — удаление текущего аккаунта. Требует `Authorization: Bearer <jwt>`. Если пользователь владеет компаниями, они тоже будут удалены вместе со связанными данными.
- `POST /companies/:id/leave` — выход из компании. Обычный участник выходит без тела запроса. Владелец обязан передать `new_owner_id`, чтобы сначала назначить нового владельца.

### Пример регистрации

```bash
curl -X POST http://localhost:8000/auth/sign-up \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "StrongPass1"
  }'
```

### Пример входа

```bash
curl -X POST http://localhost:8000/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "StrongPass1"
  }'
```

### Пример восстановления пароля

```bash
curl -X POST http://localhost:8000/auth/password/forgot \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com"
  }'
```
