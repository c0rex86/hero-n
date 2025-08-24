# Multi-stage build для оптимизации размера образа
FROM golang:1.21-alpine AS builder

# Устанавливаем необходимые пакеты для сборки
RUN apk add --no-cache git ca-certificates

# Создаем пользователя для сборки
RUN adduser -D -s /bin/sh -u 1000 appuser

# Устанавливаем рабочую директорию
WORKDIR /build

# Копируем go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Финальный stage - минимальный образ
FROM alpine:latest

# Устанавливаем сертификаты для HTTPS
RUN apk --no-cache add ca-certificates

# Создаем пользователя
RUN adduser -D -s /bin/sh -u 1000 appuser

# Копируем бинарник из builder stage
COPY --from=builder /build/main /main

# Меняем владельца
RUN chown appuser:appuser /main

# Переключаемся на непривилегированного пользователя
USER appuser

# Указываем порт
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Запускаем приложение
CMD ["/main"]
