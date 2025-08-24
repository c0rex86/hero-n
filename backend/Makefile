# Backend Makefile для HERO!N

.PHONY: help build run test clean deps docker-build docker-run

# По умолчанию показываем помощь
help:
	@echo "HERO!N Backend - команды для разработки"
	@echo ""
	@echo "Основные команды:"
	@echo "  build       - собрать бинарник"
	@echo "  run         - запустить сервер"
	@echo "  test        - запустить тесты"
	@echo "  clean       - очистить артефакты"
	@echo "  deps        - установить зависимости"
	@echo "  docker-build - собрать Docker образ"
	@echo "  docker-run  - запустить в Docker"
	@echo ""
	@echo "Разработка:"
	@echo "  dev         - запустить в режиме разработки"
	@echo "  lint        - проверить код линтером"
	@echo "  format      - отформатировать код"

# Собрать бинарник
build:
	@echo "Собираем HERO!N backend..."
	@go build -o bin/main ./cmd
	@echo "Бинарник готов: bin/main"

# Запустить сервер
run: build
	@echo "Запускаем HERO!N backend..."
	@./bin/main

# Запустить в режиме разработки (с перезапуском)
dev:
	@echo "Запускаем в режиме разработки..."
	@air

# Запустить тесты
test:
	@echo "Запускаем тесты..."
	@go test ./...

# Очистить артефакты
clean:
	@echo "Очищаем артефакты..."
	@rm -rf bin/
	@go clean

# Установить зависимости
deps:
	@echo "Устанавливаем зависимости..."
	@go mod download
	@go mod tidy

# Проверить код линтером
lint:
	@echo "Проверяем код линтером..."
	@golangci-lint run

# Отформатировать код
format:
	@echo "Форматируем код..."
	@go fmt ./...
	@gofmt -w .

# Собрать Docker образ
docker-build:
	@echo "Собираем Docker образ..."
	@docker build -t hero-n-backend .

# Запустить в Docker
docker-run:
	@echo "Запускаем в Docker..."
	@docker run -p 8080:8080 -p 4001:4001 hero-n-backend

# Создать базу данных (если используется PostgreSQL)
db-init:
	@echo "Инициализируем базу данных..."
	# TODO: добавить скрипт инициализации БД

# Миграции базы данных
db-migrate:
	@echo "Применяем миграции..."
	# TODO: добавить миграции

# Запустить с конфигом
run-with-config: build
	@echo "Запускаем с конфигурацией..."
	@./bin/main -config=./configs/config.yaml

# Показать статус
status:
	@echo "Статус HERO!N backend:"
	@echo "- Порт API: 8080"
	@echo "- Порт P2P: 4001"
	@echo "- Статус: $(shell curl -s http://localhost:8080/health 2>/dev/null && echo "запущен" || echo "не запущен")"
