# Heroin Roadmap

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

## Цели

Heroin объединяет файловое хранилище и мессенджер в одном приложении. Данные хранятся в IPFS только в зашифрованном виде. Сообщения защищены e2e. Сеть p2p на базе libp2p с адаптивной маршрутизацией и обходом блокировок. Клиент только Android. Сервер на Go выполняет роли bootstrap, ретранслятора, минимальных учетных операций и интеграции с IPFS.

## Фазы

### Фаза 0. Фундамент
- Репозиторий и структура модулей (server, client/android, shared/proto)
- Конфиг и загрузчик YAML
- SQLite, миграции, таблицы users, devices, sessions, files, messages
- Базовые proto v1: auth, messaging, storage, генерация stubs
- HTTP health и gRPC заготовки
- IPFS клиент: pin/add, dag/import, dag/export
- Observability: JSON-логи, Prometheus метрики
- Docker-compose с IPFS, Prometheus, Grafana

Статус: выполнено на базовом уровне, сервер собирается и запускается, есть health.

### Фаза 1. MVP

Цель: рабочий прототип, 1:1 чаты и безопасная передача файлов между двумя устройствами одного пользователя.

Сервер:
- Auth
  - SRP или PAKE proof проверки на сервере
  - Refresh-токены с ротацией и отзывом, аудит логинов и обновлений
  - Рейтконтроль на чувствительных RPC
- Messaging
  - Протобаф Envelope: conversation_id, message_id, ciphertext, signature, sent_at_unix
  - Подпись конвертов Ed25519, проверка и дедупликация
  - Offline очередь и Pull с курсором времени
- Storage
  - Потоковая отдача CAR из IPFS, backpressure и лимиты
  - Реальный BLAKE3 для валидации CAR до публикования
  - Политика пиннинга и ретраев, базовый репликатор
- P2P
  - DHT discovery, bootstrap список из конфига, rendezvous для устройств одного пользователя
  - Relay v2 enable, AutoNAT, AutoRelay
- Routing
  - Сбор метрик кандидатов, активный пробинг, скоринг, выбор маршрута
  - Fallback порядок: p2p, relay v2, сервер ретранслятор, офлайн
- Observability
  - Метрики по auth, storage, messaging, p2p, routing
  - Базовый OTLP-трейсинг (экспорт в Jaeger/Tempo)

Android:
- Core-crypto: XChaCha20-Poly1305, X25519, Ed25519, Argon2id, Keystore
- Core-network: gRPC клиент, менеджер транспорта, WSS и QUIC, proto модели
- Core-storage: IPFS клиент, CAR streaming, BLAKE3 проверка
- Feature-auth: экраны регистрации, логина, 2FA, хранение токенов
- Feature-messenger: список чатов, экран диалога, отправка и прием e2e
- Feature-files: список, загрузка и скачивание с прогрессом, предпросмотр
- App: DI, навигация, темы, настройки

Тесты:
- Юнит тесты crypto, auth, routing, storage hashing
- Интеграционные тесты gRPC, IPFS публикование/скачивание
- Нагрузочные для пиннинга и очереди сообщений

Выходные артефакты фазы: работающий сервер, Android APK с MVP функциональностью.

### Фаза 2. Группы и адаптивная маршрутизация

- Групповые чаты, распределение ключей на группу, вступление/выход участников
- Offline доставка с гарантиями, дедупликация и повторная отправка
- Полная адаптивная маршрутизация: миграция транспорта без разрыва, сохранение сеансов
- Обход блокировок: WSS:443, QUIC на альтернативных доменах, DoH, ротация доменов и портов, цепочки relay
- Расширенная наблюдаемость и алерты

### Фаза 3. Оптимизация и надежность

- Энергоэффективность клиента, оптимизация сетевых расходов
- Улучшение задержек и throughput, тюнинг QUIC
- Фоновая репликация пинов IPFS
- Аудиты безопасности, фуззинг, хаос-тестирование сети
- Подготовка к публичному релизу

## Подробный план по модулям

### server/internal/config
- Расширить валидацию конфигурации, дефолты и предупреждения
- Поддержка OTLP endpoint, включение/отключение метрик по модулям

### server/internal/store
- Репозитории для users, devices, sessions, files, messages с контекстными ошибками
- Тесты: миграции, CRUD, ограничения и индексы

### server/internal/auth
- Интеграция SRP или PAKE
- Переход на PASETO v2 с локальным симметричным ключом
- Refresh токены с blacklist и отзывом сессий
- Аудит: входы, смена ключей, подозрительная активность
- Рейтконтроль: фиксация IP и fingerprint клиента

### server/internal/ipfs
- Ретраи и таймауты, пул HTTP соединений
- Проверка нескольких провайдеров параллельно
- Репликация пинов по политике N

### server/internal/messaging
- Подпись и проверка конвертов, нормализация временных меток
- Идемпотентность Send, курсоры Pull, страничная выдача
- Индексы по conversation_id, sent_at_unix для быстрых выборок

### server/internal/api/{grpc,http}
- gRPC: аутентификация по токену, интерцепторы логирования, rate limiting
- HTTP: admin/metrics/debug, будущий REST шлюз при необходимости

### server/internal/bootstrap
- Создание libp2p host, DHT, AutoRelay, AutoNAT
- mDNS discovery и rendezvous канал
- Подключение к bootstrap адресам с ретраями

### server/internal/routing
- Хранение метрик скользящим окном
- Активный пробинг, SLA таймауты, расчет итоговой стоимости
- Миграция транспорта без потери e2e сеанса

### server/internal/config/observability
- OTLP экспортер, трейсинг ключевых путей: auth, messaging, storage
- Метрики p2p, routing, ipfs, БД

## Android подробный план

### client/android/core-crypto
- Обертка над libsodium, JNI, безопасные буферы
- Keystore, резервные копии ключей по фразе

### client/android/core-network
- gRPC канал с TLS, WSS, QUIC
- Менеджер транспорта, роутер, сбор метрик

### client/android/core-storage
- Клиент IPFS, загрузка и скачивание CAR с проверкой BLAKE3
- Кэширование блоков, контроль трафика

### client/android/feature-auth
- UI регистрация, логин, 2FA
- Хранение и ротация токенов

### client/android/feature-messenger
- Список чатов, диалог, отправка и прием
- Double Ratchet, перезапуск сессий

### client/android/feature-files
- Браузер файлов, предпросмотр, шаринг

### client/android/feature-settings
- Сетевые настройки, релеи, bootstrap адреса
- Управление ключами и устройствами

### client/android/app
- DI, навигация, темы, state handling, error UI

## Протоколы и версии

- v1 в `shared/proto/**` стабилизировать
- Подготовить v2 пространство имен для будущих расширений, не ломая обратную совместимость

## Тестирование

- Unit: crypto, auth, storage hashing, routing score
- Integration: gRPC, IPFS, libp2p bootstrap
- Load: пиннинг, очередь сообщений, загрузка файлов
- E2E: сценарии пользовательских операций
- Security: fuzz, static analysis, secrets scanning

## Наблюдаемость и эксплуатация

- Prometheus дашборды: auth, storage, messaging, p2p
- Логи в JSON, корреляция по request id
- Алерты: деградация маршрутов, ошибки IPFS, сбои БД

## Процессы

- Миграции БД: каждый SQL файл с номером и idempotent проверками
- Обновление proto: правка, генерация `scripts/proto-gen.sh`, сборка сервера
- Релизы: версии SemVer, Docker образы

## Критерии готовности

- Вход по username, password, 2FA, refresh токены
- Отправка и прием e2e сообщений 1:1
- Загрузка и скачивание файла через IPFS с проверкой целостности
- Автопереключение транспорта при деградации

## Backlog

- WebTransport и альтернативные домены
- Резервирование relay цепочек
- Группы и пересылка ключей
- Профилирование и оптимизация энергопотребления на Android

## Ссылки

- ТЗ: `docs/TZ.md`
- Конфиг пример: `server/configs/config.example.yaml`
- Статус: `STATUS.md`
- Гайд по контрибутингу: `CONTRIBUTING.md`
