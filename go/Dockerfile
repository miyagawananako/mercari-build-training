# syntax=docker/dockerfile:1
FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

RUN apk add --no-cache gcc musl-dev

COPY . .

RUN CGO_ENABLED=1 go build -o /app/api ./cmd/api

RUN addgroup -S mercari && adduser -S trainee -G mercari

RUN mkdir -p /app/db /app/images && \
    chown -R trainee:mercari /app/db /app/images

USER trainee

EXPOSE 9000

CMD ["/app/api"]
