# Heroin Project Status

## Обзор

Heroin — единое приложение для файлового хранилища и мессенджера с end-to-end шифрованием, хранением данных в IPFS, p2p-сетью на базе libp2p и адаптивной маршрутизацией, устойчивой к блокировкам. Сервер на Go выполняет роли bootstrap, ретранслятора и вспомогательных сервисов. Клиент только Android, UI на Jetpack Compose, архитектура модульная, ориентирована на масштабируемость и поддержку.

Основные свойства системы:
- файлы не хранятся в открытом виде на устройствах и сервере, только в шифрованном виде, публикация в IPFS через CAR
- сообщения e2e на базе X3DH и Double Ratchet, метаданные минимальны, подпись отправителя
- сеть libp2p, DHT, relay v2, AutoNAT, AutoRelay, mDNS для локального обнаружения
- транспорты QUIC как основной, TCP как fallback, WSS для обхода, WebTransport при необходимости
- маршрутизация адаптивная с активным пробингом и миграцией транспорта без разрыва сеанса
- авторизация по username и password с двойным KDF, второй фактор 10 символов с авто-ротацией
- наблюдаемость, метрики, логи, трассировка, рейтконтроль

## Что уже сделано

### Серверная часть (Go)

Архитектура и конфигурация:
- модульная структура `server/internal/*`, границы модулей стабильны
- загрузка YAML-конфига `internal/config/loader.go`, разбор `server/configs/config.example.yaml`
- observability: структурные JSON-логи, Prometheus метрики `internal/config/observability.go`

Хранилище данных:
- SQLite с WAL, подключение `internal/store/sqlite.go`
- миграции через embed.FS: `server/migrations/*.sql`
- таблицы: users, devices, sessions, files, messages, индексы и внешние ключи

Криптография и учет:
- пароли через Argon2id (параметры из конфига), двойной KDF стек
- временная реализация HMAC-токенов (перейдем на PASETO), конфигурируемый TTL
- второй фактор: код длиной 10 символов с ротацией по времени, допускается окно времени

IPFS интеграция:
- клиент IPFS HTTP API `internal/ipfs/client.go`
- публикация CAR: `dag/import`, пиннинг `pin/add`
- экспорт CAR потоками `dag/export` для отдачи пользователю

Messaging:
- очередь офлайн-конвертов `internal/messaging/queue.go` с дедупликацией по паре conversation_id, message_id
- сервис messaging `internal/messaging/service.go` принимает зашифрованные конверты, поддерживает Pull с отсечкой по времени

API:
- gRPC сервер `internal/api/grpc/server.go` на базе сгенерированных stubs в `internal/gen`
- HTTP-health `internal/api/http/server.go` для проверки живости
- Protobuf v1: `shared/proto/auth/v1`, `shared/proto/messaging/v1`, `shared/proto/storage/v1`, генерация в `server/internal/gen`

Сеть и маршрутизация:
- скелет libp2p bootstrap-узла `internal/bootstrap/node.go` с relay v2 и mDNS
- движок маршрутизации `internal/routing/engine.go`: скоринг по задержке, потерям, джиттеру, стабильности, риску блокировки и нагрузке

Инфраструктура:
- Dockerfile для сервера, docker-compose c IPFS, Prometheus, Grafana
- локальный конфиг `docker/config.local.yaml`
- скрипты `scripts/dev-start.sh`, `scripts/dev-stop.sh`, `scripts/proto-gen.sh`
- README с инструкциями и структурами

### Android (подготовка)

- каркас мульти-модульного проекта: `client/android/settings.gradle.kts`, корневой `build.gradle.kts`
- модуль `app` с Compose, Hilt, Navigation, зависимости
- заглушки модулей `core-crypto`, `core-network`, `core-storage`, `feature-*`
- README клиента с архитектурой и планом

## Подробности реализации по ТЗ

Криптография:
- симметричное шифрование: XChaCha20-Poly1305
- диффи-хеллман: X25519
- подписи: Ed25519
- хеш: BLAKE3
- KDF: Argon2id (клиент и сервер), параметры конфигурируемы
- генерация случайных значений: системный источник
- кодирование ключей: base64url

Управление ключами:
- главная пара для подписи и установления связи
- сессионные пары для диалогов
- приватные ключи только у клиента (Android Keystore)
- резервные копии ключей по инициативе пользователя, шифрованы парольной фразой

Протокол сообщений E2E:
- инициализация X3DH
- сеансовая передача Double Ratchet, прерывистый перезапуск
- защита от повторов счетчиками и оконными фильтрами
- подпись отметок времени
- минимизация метаданных

Шифрование файлов:
- уникальный ключ и nonce на файл
- потоковое шифрование файла, ключ шифруется для каждого получателя X25519
- целостность Poly1305 и кусочная валидация BLAKE3

Сеть и транспорты:
- QUIC основной, TCP fallback, WSS поверх TLS для обходов, WebTransport при поддержке
- все соединения поверх TLS, корректный fingerprint через uTLS
- селектор транспорта учитывает тип сети и историю качества

libp2p:
- DHT для поиска пиров, GossipSub для событий, Circuit Relay v2 для обхода NAT
- AutoNAT и AutoRelay, hole punching
- рандеву для встреч устройств одного пользователя

Обнаружение:
- mDNS в локальной сети, глобально через DHT и список bootstrap адресов
- обновление списка через конфиг и служебные топики

Устойчивость к блокировкам:
- маскировка под обычный TLS-трафик, WSS на 443, DoH для резолва
- ротация доменов и портов, цепочки relay при жестких ограничениях

Планировщик маршрутов:
- таблица кандидатов: p2p прямые, через реле, сервер ретранслятор, шлюзы ipfs
- метрики: задержка, потери, джиттер, разрывы, доля успеха, признаки блокировок
- выбор по минимальной стоимости, активный пробинг, миграция транспорта без разрыва сеанса
- конфигурируемые пороги переключения и веса

IPFS хранение:
- публикация: шифрование на клиенте, сборка CAR, публикация блоков, по запросу — пин и репликация на N узлов
- доступ: скачивание по cid через локальный узел или шлюз, верификация целостности, загрузка недостающих частей параллельно
- метаданные: описание, размер, mime, cid, схема шифрования, отметки времени, владельцы, получатели

Сервер Go:
- роли: bootstrap, ретрансляция, пиннинг IPFS, минимальные учетные операции
- gRPC и HTTP API
- очередь офлайн конвертов
- база данных SQLite с WAL, миграции, индексы по пользователям и cid
- журналирование аудита входов и смен ключей (запланировано)
- конфигурация YAML: порты TCP/QUIC, bootstrap адреса и реле, настройки IPFS и пиннинга, параметры KDF и сроки токенов
- наблюдаемость: JSON-логи, метрики Prometheus, OpenTelemetry трассировка (подключение OTLP в планах), рейтконтроль

Клиент Android:
- архитектура: модули ядра для сети, крипто, хранилища; фичи для чатов, файлов, авторизации, настроек
- UI: Jetpack Compose, светлая/темная тема, экраны чатов, диалог, список файлов, просмотр, загрузка, настройки устройств и ключей
- сеть: libp2p клиент или обертка над go-libp2p через gRPC и WSS; поддержка quic и tcp; менеджер транспорта и роутер в core-network
- крипто: Android Keystore, libsodium для XChaCha20-Poly1305 и X25519, безопасное хранилище сессионных ключей, авто пересоздание ратчетов
- хранилище: IPFS через локальный узел или удаленный API, кэш и предзагрузка по Wi-Fi

## Конфигурация

Файл: `server/configs/config.example.yaml`.

Ключевые секции:
- server.listen: tcp и quic адреса
- server.transports: включение tcp, quic, ws, wss
- routing: стратегия, bootstrap_nodes, relays, пороги переключений, окно метрик
- ipfs: endpoint, pinning_enabled, replication_factor
- security.kdf: тип, time, memory_mb, threads, key_len
- security.token: issuer, lifetime_min, refresh_days
- security.secondary_key: length, rotate_minutes, allowed_clock_skew_sec
- database: dsn SQLite с pragma WAL и foreign_keys
- logging: уровень
- observability: prometheus_addr, otlp_endpoint

## База данных

Таблицы:
- users: id, username, server_salt, password_hash, public_key, second_factor_secret, created_at
- devices: устройства пользователя, уникальность по user_id+device_id
- sessions: refresh-токены на устройство, TTL, индексы по user_id, device_id
- files: метаданные и cid, индексы по user_id и cid
- messages: офлайн конверты, уникальность по conversation_id+message_id

Практики:
- WAL включен, foreign_keys ON
- идempotency в insert через уникальные индексы

## Наблюдаемость

Метрики Prometheus:
- heroin_requests_total: счетчик запросов по методу, endpoint, статусу
- heroin_request_duration_seconds: гистограмма длительности
- heroin_active_connections: gauge активных соединений

Логи:
- JSON handler, уровни info/error, привязка к контекстам

Трейсинг:
- планируется OTLP-экспортер и интеграция с OpenTelemetry SDK

## Риски и меры

- блокировки протоколов. контрмеры: WSS:443, QUIC на альтернативных доменах, DoH, relay цепочки
- потеря ключей. контрмеры: резервная фраза у пользователя, шифрование
- перегрузка серверов. контрмеры: масштабирование реле и пиннеров, рейтконтроль

## План работ по модулям

Сервер:
- config: завершить валидацию, дефолты и OTLP настройки `internal/config/*`
- store: доделать репозитории для всех таблиц, подготовить интеграционные тесты `internal/store/*`
- auth: реализовать SRP или PAKE, перейти на PASETO v2, добавить refresh-токены, аудит `internal/auth/service.go`
- ipfs: добавить репликацию и проверку доступности нескольких провайдеров `internal/ipfs/client.go`
- messaging: расширить Envelope до protobuf-представления, подписи и проверка повторов `internal/messaging/*`
- api: полноценно покрыть gRPC и HTTP, auth middleware, рейтконтроль `internal/api/{grpc,http}`
- p2p/bootstrap: DHT, AutoRelay, AutoNAT, rendezvous `internal/bootstrap/node.go`, `internal/discovery/*`
- routing: активный пробинг, скользящее окно метрик, миграция транспорта без разрыва `internal/routing/engine.go`
- observability: OTLP трассировка, метрики для p2p и ipfs `internal/config/observability.go`

Протоколы:
- уточнить поля и статусы в `shared/proto/**`, обеспечить версионирование v1 и место для v2
- автоматизировать генерацию stubs `scripts/proto-gen.sh`

Android:
- core-crypto: XChaCha20-Poly1305, X25519, Ed25519, Argon2id, Keystore `client/android/core-crypto`
- core-network: libp2p, QUIC/TCP/WSS, менеджер транспорта и роутер `client/android/core-network`
- core-storage: IPFS клиент, CAR streaming, BLAKE3 верификация `client/android/core-storage`
- feature-*: auth, messenger, files, settings на Compose `client/android/feature-*`
- app: навигация, темы, DI `client/android/app`

Маршрутизация подробно:
- кандидаты: прямое p2p, relay v2, сервер ретранслятор, офлайн доставка
- метрики: задержка, потери, джиттер, стабильность, риск блокировки, загрузка
- активный пробинг и скользящее окно
- выбор маршрута по стоимости, бесшовная миграция транспорта
- обход блокировок через wss:443, quic на альтернативных доменах, ротация доменов и портов, DoH
- пороги переключений и веса конфигурируемы

Безопасность:
- двойной KDF, проверка повторов, подпись конвертов, минимизация метаданных
- журнал аудита входов и смен ключей
- рейтконтроль на API и защита от перегрузки

## Как запустить локально

Предварительно: Go 1.22+, Docker и Docker Compose, protoc. Команда запуска:

```bash
./scripts/dev-start.sh
```

Сервисы:
- Heroin gRPC: localhost:8080
- Heroin HTTP: localhost:8082
- IPFS API: localhost:5001, Gateway: localhost:8081
- Prometheus: localhost:9091, Grafana: localhost:3000

Проверка:

```bash
curl http://localhost:8082/healthz
```

Остановка:

```bash
./scripts/dev-stop.sh
```

## Текущий статус

Готовность на данный момент: около 2 процентов. Сервер имеет базовую функциональность: конфиг, БД с миграциями, IPFS публикация и экспорт, очередь сообщений, gRPC и HTTP точки входа, метрики. Android подготовлен структурно, требуется реализация модулей. P2P собран частично, предстоит интеграция DHT, AutoRelay и маршрутизатора. E2E протоколы описаны в ТЗ, реализации еще нет.

Готов к тестированию: Auth API (базовый), Storage upload/download через IPFS, локальное dev окружение.

Следующий крупный milestone: рабочий MVP с базовым auth, обменом e2e сообщениями один-на-один, загрузкой и скачиванием файлов через IPFS, автоматическим переключением транспорта при деградации, и минимальными экранами на Android.
