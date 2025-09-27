# Heroin - Secure P2P File Storage & Messaging

### Важное примечание для разработчиков

- **ИИ приветствуется как инструмент для рутинных задач и частичного написания кода**, но:
  - **вы должны понимать что делаете** — ИИ может ошибаться, проверяйте результат
  - **вы должны понимать работу кода** — не добавляйте код, который не можете объяснить
  - **вы должны уметь писать код сами** — ИИ это помощник, не замена
  - **вы должны оставлять комментарии** — для навигации и объяснения сложных мест

- **Почему это важно**
  - код должен быть понятен и поддерживаем
  - безопасность и надежность важнее скорости
  - ответственность за внесенный код на разработчике
  - комментарии ускоряют чтение и ревью

- **Рекомендации**
  - используйте ИИ для шаблонов и рутины
  - всегда проверяйте сгенерированное
  - добавляйте поясняющие комментарии
  - пишите код понятным без ИИ

[![Go](https://img.shields.io/badge/Go-1.22%2B-blue.svg)](https://golang.org) [![Kotlin](https://img.shields.io/badge/Kotlin-1.9%2B-purple.svg)](https://kotlinlang.org) [![LibP2P](https://img.shields.io/badge/LibP2P-0.32%2B-green.svg)](https://libp2p.io) [![IPFS](https://img.shields.io/badge/IPFS-Kubo%200.24%2B-yellow.svg)](https://ipfs.tech) [![gRPC](https://img.shields.io/badge/gRPC-Ready-orange.svg)](https://grpc.io) [![Protobuf](https://img.shields.io/badge/Protobuf-v21%2B-lightgrey.svg)](https://protobuf.dev) [![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://docker.com) [![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE) [![PRs Welcome](https://img.shields.io/badge/PRs-Welcome-brightgreen.svg)](CONTRIBUTING.md)

### Технологический стек

- **Backend (Go)**
  - Go 1.22+, gRPC, Protobuf
  - SQLite (modernc.org/sqlite), миграции через embed FS
  - libp2p 0.32.x: DHT, GossipSub, Circuit Relay v2, AutoNAT, AutoRelay, mDNS, rendezvous
  - Транспорты: QUIC (quic-go), TCP, WS/WSS, HTTP/2, опционально WebTransport
  - Observability: Prometheus, структурные JSON-логи, OpenTelemetry (OTLP)
  - Конфигурация: YAML
  - Контейнеры: Docker, docker-compose

- **Crypto**
  - Argon2id (двойной KDF), HMAC
  - XChaCha20-Poly1305, X25519, Ed25519
  - BLAKE3 для верификации файлов (на стороне сервера и клиента)
  - base64url кодирование

- **Storage / IPFS**
  - IPFS Kubo API: `dag/import`, `dag/export`, `pin/add`
  - CAR streaming, пиннинг, репликация на N узлов

- **Networking / Routing**
  - Адаптивная маршрутизация: сбор метрик, скоринг, активный пробинг, миграция транспорта
  - Обход блокировок: WSS:443, QUIC на альтернативных доменах, DoH, ротация доменов и портов, цепочки relay
  - uTLS fingerprint

- **Android**
  - Kotlin 1.9+, Jetpack Compose, Material 3, Navigation
  - Hilt, Coroutines, Flow
  - gRPC, OkHttp
  - libsodium, Android Keystore

- **Инфраструктура и процессы**
  - Скрипты: `scripts/proto-gen.sh`, `scripts/dev-start.sh`, `scripts/dev-stop.sh`
  - Документация: `README.md`, `STATUS.md`, `ROADMAP.md`, `CONTRIBUTING.md`
  - Версионирование: SemVer, Docker образы сервера

Единое приложение для файлового хранилища и мессенджера с сквозным шифрованием, IPFS хранением и устойчивой P2P маршрутизацией.

## Быстрый старт

### Требования

- Go 1.22+
- Docker & Docker Compose
- protoc (для генерации gRPC)

### Запуск development окружения

```bash
# клонировать и перейти в директорию
git clone <repo-url> heroin
cd heroin

# запустить все сервисы
./scripts/dev-start.sh

# проверить статус
curl http://localhost:8082/healthz
```

### Сервисы

- **Heroin Server**: 
  - gRPC: `localhost:8080`
  - HTTP: `localhost:8082`
  - Metrics: `localhost:9090`
- **IPFS**: 
  - API: `localhost:5001`
  - Gateway: `localhost:8081`
- **Мониторинг**:
  - Prometheus: `localhost:9091`
  - Grafana: `localhost:3000` (admin/admin)

### Остановка

```bash
./scripts/dev-stop.sh
```

## Разработка

### Структура проекта

```
heroin/
├── server/                 # Go сервер
│   ├── cmd/heroin-server/ # Точка входа
│   ├── internal/          # Внутренние пакеты
│   ├── configs/           # Конфигурации
│   └── migrations/        # SQL миграции
├── client/android/        # Android клиент
├── shared/proto/          # Protobuf схемы
├── docker/               # Docker файлы
└── scripts/              # Утилиты
```

### Генерация protobuf

```bash
./scripts/proto-gen.sh
```

### Сборка сервера

```bash
cd server
go build ./cmd/heroin-server
```

### API Endpoints

#### gRPC Services

- `auth.v1.AuthService` - регистрация, авторизация
- `messaging.v1.MessagingService` - отправка сообщений
- `storage.v1.StorageService` - файловое хранилище

#### HTTP

- `GET /healthz` - проверка здоровья

## Конфигурация

Основной файл: `server/configs/config.example.yaml`

Ключевые параметры:
- `server.listen` - адреса для TCP/QUIC
- `ipfs.endpoint` - IPFS API endpoint
- `security.kdf` - параметры Argon2id
- `routing` - настройки P2P маршрутизации

## Безопасность

- **Пароли**: двойной KDF с Argon2id
- **Токены**: HMAC-подписанные с ротацией
- **2FA**: TOTP-подобные коды с 30-минутной ротацией
- **Файлы**: XChaCha20-Poly1305 + Blake3 верификация
- **Сообщения**: Double Ratchet E2E шифрование

## Мониторинг

### Метрики Prometheus

- `heroin_requests_total` - общее количество запросов
- `heroin_request_duration_seconds` - время обработки
- `heroin_active_connections` - активные соединения

### Логи

Структурированные JSON логи с уровнями info/error.

## Android клиент

### Модули

- `core-crypto` - криптографические примитивы
- `core-network` - P2P сеть и маршрутизация  
- `core-storage` - IPFS интеграция
- `feature-auth` - авторизация
- `feature-messenger` - чаты
- `feature-files` - файлы
- `feature-settings` - настройки

### Сборка

```bash
cd client/android
./gradlew build
```

## Развертывание

### Production

1. Настроить TLS сертификаты
2. Обновить bootstrap узлы в конфиге
3. Настроить мониторинг и алерты
4. Запустить через docker-compose

### Масштабирование

- Несколько bootstrap узлов для отказоустойчивости
- Реплики IPFS пиннеров
- Load balancer для gRPC/HTTP

## Разработка

### Добавление новых API

1. Обновить proto файлы в `shared/proto/`
2. Запустить `./scripts/proto-gen.sh`
3. Реализовать методы в соответствующих сервисах
4. Добавить тесты

### Миграции БД

Создать файл `server/migrations/XXX_description.sql` с SQL командами.

## Лицензия

MIT
