FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/migrate ./cmd/migrate

FROM alpine:3.20

WORKDIR /app

COPY --from=build /bin/server /app/server
COPY --from=build /bin/migrate /app/migrate
COPY migrations /app/migrations

EXPOSE 8000

CMD ["./server"]
