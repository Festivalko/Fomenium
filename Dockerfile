# Используем официальный образ Go 1.25.8
FROM golang:1.25.8-alpine AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off
ENV GOTOOLCHAIN=local

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарники
RUN go build -o validator cmd/validator/main.go
RUN go build -o multiplexer cmd/multiplexer/main.go
RUN go build -o client cmd/client/main.go
RUN go build -o demo cmd/demo/main.go

# Финальный образ
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/validator /app/validator
COPY --from=builder /app/multiplexer /app/multiplexer
COPY --from=builder /app/client /app/client
COPY --from=builder /app/demo /app/demo

RUN chmod +x /app/*

EXPOSE 8001 8002 8080

CMD ["/app/validator"]
