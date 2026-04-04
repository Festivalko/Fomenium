FROM alpine:latest

RUN apk add --no-cache ca-certificates curl

WORKDIR /app

COPY bin/validator /app/validator
COPY bin/multiplexer /app/multiplexer
COPY bin/client /app/client
COPY bin/demo /app/demo

RUN chmod +x /app/*

# Устанавливаем supervisord для запуска нескольких процессов
RUN apk add --no-cache supervisor
COPY supervisord.conf /etc/supervisord.conf

EXPOSE 8001 8002 8080

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
