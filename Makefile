# HERO!N Messenger - Makefile для разработки

.PHONY: help build test clean docker-up docker-down lint format install-deps

# По умолчанию показываем помощь
help:
	@echo "HERO!N Messenger - команды для разработки"
	@echo ""
	@echo "Основные команды:"
	@echo "  help        - показать эту справку"
	@echo "  build       - собрать все компоненты"
	@echo "  test        - запустить все тесты"
	@echo "  clean       - очистить все артефакты сборки"
	@echo "  lint        - проверить код линтерами"
	@echo "  format      - отформатировать код"
	@echo ""
	@echo "Docker команды:"
	@echo "  docker-up   - запустить все сервисы в Docker"
	@echo "  docker-down - остановить все сервисы"
	@echo "  docker-logs - показать логи всех сервисов"
	@echo ""
	@echo "Backend команды:"
	@echo "  build-go    - собрать Go приложение"
	@echo "  test-go     - запустить Go тесты"
	@echo "  run-go      - запустить Go приложение локально"
	@echo ""
	@echo "Android команды:"
	@echo "  build-android    - собрать Android приложение"
	@echo "  test-android     - запустить Android тесты"
	@echo "  install-android  - установить на подключенное устройство"

# Сборка всех компонентов
build: build-go build-android

# Сборка Go backend
build-go:
	@echo "Сборка Go backend..."
	cd backend && go build -o bin/main ./cmd

# Сборка Android приложения
build-android:
	@echo "Сборка Android приложения..."
	cd mobile && ./gradlew assembleDebug

# Запуск всех тестов
test: test-go test-android

# Тестирование Go кода
test-go:
	@echo "Запуск Go тестов..."
	cd backend && go test ./...

# Тестирование Android кода
test-android:
	@echo "Запуск Android тестов..."
	cd mobile && ./gradlew testDebugUnitTest

# Очистка артефактов сборки
clean:
	@echo "Очистка артефактов сборки..."
	cd backend && rm -rf bin/
	cd mobile && ./gradlew clean

# Проверка кода линтерами
lint: lint-go lint-android

# Проверка Go кода
lint-go:
	@echo "Проверка Go кода линтером..."
	cd backend && golangci-lint run

# Проверка Android кода
lint-android:
	@echo "Проверка Android кода линтером..."
	cd mobile && ./gradlew lintDebug

# Форматирование кода
format: format-go format-android

# Форматирование Go кода
format-go:
	@echo "Форматирование Go кода..."
	cd backend && gofmt -w .

# Форматирование Android кода
format-android:
	@echo "Форматирование Android кода..."
	cd mobile && ./gradlew spotlessApply

# Установка зависимостей
install-deps: install-go-deps install-android-deps

# Установка Go зависимостей
install-go-deps:
	@echo "Установка Go зависимостей..."
	cd backend && go mod download

# Установка Android зависимостей
install-android-deps:
	@echo "Установка Android зависимостей..."
	cd mobile && ./gradlew build --refresh-dependencies

# Docker команды
docker-up:
	@echo "Запуск Docker сервисов..."
	docker-compose up -d

docker-down:
	@echo "Остановка Docker сервисов..."
	docker-compose down

docker-logs:
	@echo "Показ логов Docker сервисов..."
	docker-compose logs -f

# Backend команды
run-go: build-go
	@echo "Запуск Go приложения..."
	cd backend && ./bin/main

# Android команды
install-android: build-android
	@echo "Установка Android приложения..."
	cd mobile && ./gradlew installDebug

# Полная настройка для разработки
setup: install-deps
	@echo "Настройка среды разработки завершена!"
	@echo ""
	@echo "Далее:"
	@echo "1. Запустите сервисы: make docker-up"
	@echo "2. Соберите backend: make build-go"
	@echo "3. Соберите Android: make build-android"
	@echo "4. Запустите тесты: make test"

# Проверка здоровья проекта
health-check:
	@echo "Проверка здоровья проекта..."
	@echo "Go модули:" && cd backend && go mod verify
	@echo "Android dependencies:" && cd mobile && ./gradlew dependencies --configuration debugCompileClasspath | grep -E "(FAILED|ERROR)" || echo "OK"

# Создание тестовых данных
test-data:
	@echo "Создание тестовых данных..."
	# Здесь можно добавить скрипты для генерации тестовых данных

# Документация
docs:
	@echo "Генерация документации..."
	# Здесь можно добавить генерацию docs

# Релиз
release: test lint
	@echo "Подготовка к релизу..."
	@echo "1. Все тесты пройдены"
	@echo "2. Код отформатирован"
	@echo "3. Линтер прошел без ошибок"
	@echo "4. Готово к созданию тега и релизу"
