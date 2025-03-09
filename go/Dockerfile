# syntax=docker/dockerfile:1
FROM golang:1.24-alpine

ENV CGO_ENABLED=1

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

RUN apk add --no-cache gcc musl-dev

COPY . .

RUN addgroup -S mercari && adduser -S trainee -G mercari

RUN mkdir -p /app/db /app/images && \
    chown -R trainee:mercari /app/db /app/images

USER trainee

CMD ["go", "run", "cmd/api/main.go"]
