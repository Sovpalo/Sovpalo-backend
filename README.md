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

## Деплой на сервер

Скрипт `scripts/deploy.sh` подключается к серверу по SSH, при необходимости устанавливает Docker, клонирует или обновляет репозиторий, пересобирает контейнеры, применяет миграции и запускает API.

```bash
./scripts/deploy.sh
```

По умолчанию деплой идет на `root@2.56.241.112` в `/opt/sovpalo/app` из ветки `main`. При первом запуске на сервере будет создан `.env` из `.env.example`; после этого нужно один раз заполнить реальные секреты в `/opt/sovpalo/app/.env`.

Параметры можно переопределить переменными окружения:

```bash
BRANCH=main APP_DIR=/opt/sovpalo/app SSH_KEY=~/.ssh/id_ed25519 ./scripts/deploy.sh
```

## Миграции

Команды запускаются через `cmd/migrate` (используется goose):

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down
go run ./cmd/migrate status
```

## Архитектура аутентификации

Сервис использует две сущности:

- `users` — профиль пользователя (`username`, `avatar_url` и другие пользовательские данные).
- `auth_identities` — способы входа в аккаунт (`password`, `telegram`).

Это позволяет:

- держать `email/password` как основной способ входа;
- подключать Telegram как дополнительный способ;
- создавать пользователей без `email`;
- позже безопасно добавить привязку нескольких способов входа к одному аккаунту.

## Эндпоинты

- `GET /health` — проверка доступности сервиса и базы данных.
- `GET /health/smtp` — проверка SMTP-подключения и SMTP-аутентификации.
- `POST /auth/sign-up` — начало регистрации. Принимает `username`, `email`, `password`, отправляет 4-значный код на email.
- `POST /auth/sign-up/verify` — подтверждение кода. Принимает `email`, `code`, создаёт пользователя и возвращает JWT.
- `POST /auth/sign-up/resend` — повторная отправка нового 4-значного кода на email.
- `POST /auth/sign-in` — вход по `email` и `password`, сразу возвращает JWT.
- `POST /auth/telegram/sign-in` — вход через Telegram Login Widget или WebApp init data. Для Login Widget принимает поля Telegram-пользователя (`id`, `first_name`, `auth_date`, `hash`, опционально `last_name`, `username`, `photo_url`). Для Mini App/WebApp принимает `init_data` — сырую строку `Telegram.WebApp.initData`. Возвращает JWT. Для Telegram-пользователя `email` в профиле может отсутствовать.
- `POST /auth/password/forgot` — запуск восстановления пароля по `email`, отправляет 4-значный код на email.
- `POST /auth/password/verify` — подтверждение кода и установка нового пароля. Принимает `email`, `code`, `new_password`.
- `POST /auth/password/resend` — повторная отправка кода для восстановления пароля.
- `GET /auth/me` — получение информации о текущем пользователе. Требует `Authorization: Bearer <jwt>`, возвращает `username`, `avatar_url`, опциональный `email` и список `providers`.
- `POST /auth/me/avatar` — загрузка аватарки текущего пользователя. Требует `Authorization: Bearer <jwt>` и `multipart/form-data` с полем `avatar`. Поддерживаются PNG/JPEG/WEBP/GIF до 5 MB.
- `DELETE /auth/me/avatar` — удаление аватарки текущего пользователя. Требует `Authorization: Bearer <jwt>`.
- `DELETE /auth/me` — удаление текущего аккаунта. Требует `Authorization: Bearer <jwt>`. Если пользователь владеет компаниями, они тоже будут удалены вместе со связанными данными.
- `POST /auth/me/push-tokens` — регистрация APNs device token текущего пользователя. Принимает `token`, `platform: "ios"`.
- `DELETE /auth/me/push-tokens` — удаление APNs device token текущего пользователя. Принимает `token`.
- `POST /companies/:id/leave` — выход из компании. Обычный участник выходит без тела запроса. Владелец обязан передать `new_owner_id`, чтобы сначала назначить нового владельца.
- `POST /companies` — создание компании. Принимает `name`, опционально `description` и `avatar_url`.
- `PATCH /companies/:id` — обновление компании владельцем. Поддерживает `application/json` с `name`, `description`, `avatar_url` и `multipart/form-data` с полями `name`, `description`, `avatar_url`, `avatar`. Файл `avatar` сохраняется на сервере, а в `avatar_url` записывается URL.
- `POST /events` и `POST /companies/:id/events` — создание встречи. Поддерживают `application/json` с `photo_url` и `multipart/form-data` с полями `title`, `description`, `photo_url`, `start_time`, `end_time`, `company_id`, `photo`. Файл `photo` сохраняется на сервере, а в `photo_url` записывается URL.
- `PATCH /events/:id` и `PATCH /companies/:id/events/:event_id` — обновление встречи. Поддерживают `application/json` с `photo_url` и `multipart/form-data` с полями `title`, `description`, `photo_url`, `start_time`, `end_time`, `photo`. Файл `photo` сохраняется на сервере, а в `photo_url` записывается URL.
- `POST /companies/:id/ideas` — создание идеи. Поддерживает `application/json` с `photo_url` и `multipart/form-data` с полями `title`, `description`, `photo_url`, `photo`. Файл `photo` сохраняется на сервере, а в `photo_url` записывается URL.
- `POST /companies/:id/ideas/generate` — генерация черновиков идей через YandexGPT. Принимает `topic`, опционально `context`, `audience`, `constraints`, `tone`, `count`, возвращает массив черновиков с `title`, `description`, `source`, `llm_prompt`.
- `PATCH /companies/:id/ideas/:idea_id` — обновление идеи её автором. Поддерживает `application/json` с `title`, `description`, `photo_url` и `multipart/form-data` с полями `title`, `description`, `photo_url`, `photo`. Файл `photo` сохраняется на сервере, а в `photo_url` записывается URL.
- `GET /companies/:id/chat/messages` — получить историю чата компании. Поддерживает `before_id` и `limit` для пагинации.
- `POST /companies/:id/chat/messages` — отправить сообщение в чат компании. Поддерживает `application/json` с `text` или `multipart/form-data` с одним или несколькими полями `media` для фото/видео.
- `DELETE /companies/:id/chat/messages/:message_id` — удалить своё сообщение из чата компании.
- `POST /companies/:id/chat/messages/read` — отметить сообщения как прочитанные. Принимает массив `message_ids`.
- `GET /companies/:id/chat/unread-count` — получить количество непрочитанных сообщений в чате компании.
- `GET /companies/:id/chat/ws?token=<jwt>` — WebSocket-подключение для realtime-событий `message_created`, `messages_read`, `message_deleted`.
- Push-уведомления отправляются через APNs для сообщений в чате, назначения встречи и изменения времени/места встречи. Для включения отправки задайте `APNS_KEY_ID`, `APNS_TEAM_ID`, `APNS_BUNDLE_ID` и `APNS_PRIVATE_KEY_PATH` или `APNS_PRIVATE_KEY`. Без этих переменных уведомления сохраняются в `notifications`, но push в APNs не отправляется. Для отладки: `APNS_DEBUG=true` — в логах API: endpoint (sandbox/production), `apns-topic`, Key ID, Team ID, декодированные header/payload провайдерского JWT при перевыпуске; строка `push dispatch` (получатель `recipient_user_id`, тип уведомления, число токенов, суффиксы hex); строка `push apns result` на каждый HTTP-ответ APNs (статус, `apns_reason` из JSON, суффикс device token). После отладки выключите флаг.
- Ответы со списками участников, приглашений, посещаемости, идей и доступности включают `avatar_url` пользователя там, где возвращаются данные пользователя.

Краткий контракт для мобильного клиента с JSON-примерами: [docs/mobile-chat-contract.md](/Users/gaane/dev/sovpalo-backend-clean/docs/mobile-chat-contract.md:1).
Рекомендации по mobile auth и Telegram flow: [docs/mobile-auth.md](/Users/gaane/dev/sovpalo-backend-clean/docs/mobile-auth.md:1).

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

### Пример входа через Telegram

Login Widget payload:

```bash
curl -X POST http://localhost:8000/auth/telegram/sign-in \
  -H "Content-Type: application/json" \
  -d '{
    "id": 123456789,
    "first_name": "Alice",
    "username": "alice_tg",
    "auth_date": 1710000000,
    "hash": "telegram-generated-hash"
  }'
```

Mini App/WebApp init data:

```bash
curl -X POST http://localhost:8000/auth/telegram/sign-in \
  -H "Content-Type: application/json" \
  -d '{
    "init_data": "query_id=...&user=...&auth_date=1710000000&hash=telegram-generated-hash"
  }'
```

### Пример `GET /auth/me`

```json
{
  "email": "alice@example.com",
  "username": "alice",
  "avatar_url": "/uploads/avatars/user-1-123-avatar.png",
  "providers": ["password", "telegram"]
}
```

### Пример восстановления пароля

```bash
curl -X POST http://localhost:8000/auth/password/forgot \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com"
  }'
```
