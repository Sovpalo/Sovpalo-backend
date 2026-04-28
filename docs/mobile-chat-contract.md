# Mobile Chat Contract

## Base Flow

1. Load history with `GET /companies/:id/chat/messages`.
2. Open `WebSocket` on `GET /companies/:id/chat/ws?token=<jwt>`.
3. Send text with JSON `POST /companies/:id/chat/messages`.
4. Send media with `multipart/form-data` `POST /companies/:id/chat/messages`.
5. Mark visible messages as read with `POST /companies/:id/chat/messages/read`.
6. Update badge with `GET /companies/:id/chat/unread-count`.

## HTTP Endpoints

### Get Messages

`GET /companies/:id/chat/messages?before_id=123&limit=50`

Response:

```json
[
  {
    "id": 5,
    "company_id": 9001,
    "sender_id": 9001,
    "sender_username": "chat_a",
    "sender_avatar_url": "/uploads/avatars/user-1.png",
    "text": "hello",
    "attachments": [],
    "created_at": "2026-04-28T17:03:48.488604Z",
    "is_read_by_current_user": false
  }
]
```

### Send Text Message

`POST /companies/:id/chat/messages`

```json
{
  "text": "hello"
}
```

Response:

```json
{
  "id": 5,
  "company_id": 9001,
  "sender_id": 9001,
  "sender_username": "chat_a",
  "text": "hello",
  "attachments": [],
  "created_at": "2026-04-28T17:03:48.488604Z",
  "is_read_by_current_user": true,
  "read_at": "2026-04-28T17:03:48.504975Z"
}
```

### Send Media Message

`POST /companies/:id/chat/messages`

`multipart/form-data`:
- `media`: one or more files

Response:

```json
{
  "id": 6,
  "company_id": 9001,
  "sender_id": 9001,
  "sender_username": "chat_a",
  "attachments": [
    {
      "id": 2,
      "message_id": 6,
      "file_name": "photo.png",
      "file_url": "/uploads/chat/company-9001-user-9001-1-photo.png",
      "file_type": "image/png",
      "file_size": 66,
      "media_type": "photo",
      "created_at": "2026-04-28T17:10:00.000000Z"
    }
  ],
  "created_at": "2026-04-28T17:10:00.000000Z",
  "is_read_by_current_user": true,
  "read_at": "2026-04-28T17:10:00.000000Z"
}
```

### Delete Message

`DELETE /companies/:id/chat/messages/:message_id`

Response:

```json
{
  "status": "ok"
}
```

### Mark Messages Read

`POST /companies/:id/chat/messages/read`

```json
{
  "message_ids": [5, 6, 7]
}
```

Response:

```json
{
  "message_ids": [5, 6, 7],
  "read_at": "2026-04-28T17:12:00.000000Z",
  "unread_count": 0
}
```

### Get Unread Count

`GET /companies/:id/chat/unread-count`

Response:

```json
{
  "unread_count": 3
}
```

## WebSocket Events

### message_created

```json
{
  "type": "message_created",
  "company_id": 9001,
  "message": {
    "id": 5,
    "company_id": 9001,
    "sender_id": 9001,
    "sender_username": "chat_a",
    "text": "hello",
    "attachments": [],
    "created_at": "2026-04-28T17:03:48.488604Z",
    "is_read_by_current_user": false
  }
}
```

### messages_read

```json
{
  "type": "messages_read",
  "company_id": 9001,
  "user_id": 9002,
  "message_ids": [5, 6, 7],
  "unread_count": 0,
  "read_at": "2026-04-28T17:12:00.000000Z"
}
```

### message_deleted

```json
{
  "type": "message_deleted",
  "company_id": 9001,
  "message_id": 7
}
```

## Client Rules

- A message contains either `text` or `attachments`.
- `attachments[].media_type` is `photo` or `video`.
- Sender receives created message as already read.
- Other participants receive created message as unread.
- Non-members receive `403 Forbidden`.
