# 🚀 Hero!n Messenger - Roadmap

## 📋 Обзор проекта

**Hero!n** - децентрализованный P2P мессенджер с E2EE, вдохновленный принципами торрентов. Проект ориентирован на приватность, децентрализацию и поддержку сообщества.

### 🎯 Ключевые особенности
- **P2P/E2EE** - прямое peer-to-peer общение с end-to-end шифрованием
- **I2P интеграция** - анонимизация трафика через Invisible Internet Project
- **WebRTC/QUIC** - современные протоколы для реального времени
- **IPFS хранение** - децентрализованное хранение файлов
- **DHT сеть** - распределенная система обнаружения узлов
- **STUN/TURN** - обход NAT для P2P соединений

---

## 🏗️ Архитектура системы

### 1.1 Ядро системы
```
┌─────────────────────────────────────────────────────────────┐
│                    Hero!n Messenger Core                    │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   P2P       │ │   E2EE      │ │   DHT       │           │
│  │   Engine    │ │   Crypto    │ │   Discovery │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   WebRTC    │ │   QUIC      │ │   I2P       │           │
│  │   Transport │ │   Protocol  │ │   Network   │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   IPFS      │ │   STUN      │ │   TURN      │           │
│  │   Storage   │ │   Server    │ │   Server    │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 Компоненты

#### 1.2.1 Backend (Go)
- **hero-core** - основное ядро P2P
- **hero-crypto** - криптография и E2EE
- **hero-network** - сетевая логика (I2P, STUN, DHT)
- **hero-storage** - IPFS интеграция
- **hero-api** - REST/gRPC API для клиентов

#### 1.2.2 Клиентские приложения
- **hero-desktop** - desktop клиент (Qt/Tauri)
- **hero-mobile** - мобильный клиент (React Native)
- **hero-web** - веб-клиент (React/Vue)
- **hero-cli** - консольный клиент

---

## 🔐 Безопасность и криптография

### 2.1 E2EE схема
```
Отправитель → AES-256-GCM → Получатель
             ↓
        Ed25519 подпись
             ↓
      X25519 key exchange
```

### 2.2 Ключевые особенности
- **Perfect Forward Secrecy** - новые ключи для каждой сессии
- **Post-compromise security** - защита от компрометации ключей
- **Denial of service protection** - защита от спам-атак
- **Metadata minimization** - минимальные метаданные

### 2.3 Аутентификация
- **Ed25519** для цифровых подписей
- **X25519** для key exchange (ECDH)
- **HKDF** для генерации ключей
- **BLAKE3** для хэширования

---

## 🌐 Сетевая архитектура

### 3.1 P2P сеть

#### 3.1.1 DHT (Distributed Hash Table)
```
Node ID: hash(public_key)
Buckets: Kademlia-like routing
Bootstrap: hardcoded + dynamic discovery
```

#### 3.1.2 Соединения
- **Direct P2P** - прямая связь между пирами
- **NAT traversal** - STUN/TURN для обхода NAT
- **Relay fallback** - TURN серверы для сложных случаев

### 3.2 I2P интеграция
```
Hero!n Client → I2P Router → I2P Network → Destination
```

#### 3.2.1 Преимущества
- **Полная анонимизация** - скрытые сервисы
- **Гарантированная анонимность** - встроенная в протокол
- **Защита от DPI** - трафик выглядит как шум

### 3.3 Протоколы транспорта

#### 3.3.1 WebRTC
- **DataChannel** - для сообщений и файлов
- **SRTP** - защищенный RTP
- **ICE/STUN/TURN** - NAT traversal

#### 3.3.2 QUIC
- **UDP-based** - более эффективный чем TCP
- **Built-in security** - TLS 1.3
- **Connection migration** - перенос соединений

---

## 📁 Хранение данных

### 4.1 IPFS интеграция
```
Файл → Шифрование → IPFS Hash → DHT публикация
```

#### 4.1.1 Архитектура хранения
- **Клиентское шифрование** - файлы шифруются локально
- **IPFS pinning** - закрепление важных файлов
- **Distributed storage** - распределенное хранение
- **Content addressing** - адресация по содержимому

#### 4.1.2 Управление файлами
- **Автоматическое шифрование** - перед загрузкой
- **Дедупликация** - обнаружение дубликатов
- **Версионирование** - история изменений
- **Garbage collection** - очистка неиспользуемых файлов

### 4.2 Локальное хранение
- **SQLite** - для метаданных и кэша
- **BadgerDB** - для больших объемов данных
- **BoltDB** - для конфигурации

---

## 👥 Сообщество и поддержка

### 5.1 Децентрализованная поддержка
```
Сообщество ↔ DHT сеть ↔ Bootstrap узлы ↔ Релеи
```

#### 5.1.1 Роли участников
- **Seed nodes** - начальные узлы сети
- **Relay servers** - прокси для сложных соединений
- **Storage providers** - предоставляют IPFS storage
- **Bridge operators** - мосты в другие сети

#### 5.1.2 Инфраструктура сообщества
- **Общественные TURN серверы**
- **IPFS pinning сервисы**
- **Мониторинг сети**
- **Документация и поддержка**

### 5.2 Экономика
- **Добровольные пожертвования** - для инфраструктуры
- **Token incentives** - поощрение участников
- **Staking** - для важных узлов
- **Community governance** - управление проектом

---

## 📱 Клиентские приложения

### 6.1 Desktop клиент
```
Tauri + React + Rust core
├── Native performance
├── Cross-platform
└── System integration
```

### 6.2 Мобильный клиент
```
React Native + Native modules
├── iOS/Android support
├── Push notifications
└── Background sync
```

### 6.3 Веб-клиент
```
React + WebRTC
├── Browser compatibility
├── Progressive Web App
└── Service Worker
```

### 6.4 CLI клиент
```
Go CLI application
├── Automation scripts
├── Bot integration
└── System monitoring
```

---

## 🚀 Этапы разработки

### Фаза 1: MVP (3 месяца)
```
Week 1-4: Core P2P + Basic crypto
Week 5-8: WebRTC implementation
Week 9-12: Desktop + Web clients
```

#### 1.1 Базовая P2P сеть
- [ ] DHT discovery
- [ ] Basic messaging
- [ ] Simple encryption
- [ ] Desktop client MVP

#### 1.2 Минимальный функционал
- [ ] Текстовые сообщения
- [ ] Список контактов
- [ ] Базовый UI
- [ ] Настройки профиля

### Фаза 2: Core Features (4 месяца)
```
Month 4-6: E2EE + File sharing
Month 7-8: I2P integration
```

#### 2.1 Расширенная криптография
- [ ] Полная E2EE реализация
- [ ] Forward secrecy
- [ ] Key rotation
- [ ] Secure file transfer

#### 2.2 Файловый обмен
- [ ] IPFS интеграция
- [ ] Зашифрованное хранение
- [ ] Прогресс загрузки
- [ ] Метаданные файлов

### Фаза 3: Advanced Features (3 месяца)
```
Month 9-11: Mobile + QUIC + I2P
```

#### 3.1 Мобильная платформа
- [ ] React Native клиент
- [ ] Push notifications
- [ ] Background sync
- [ ] Mobile optimizations

#### 3.2 Продвинутые протоколы
- [ ] QUIC implementation
- [ ] I2P tunneling
- [ ] Advanced NAT traversal
- [ ] Connection migration

### Фаза 4: Production Ready (2 месяца)
```
Month 12-13: Testing + Community
```

#### 4.1 Качество и тестирование
- [ ] End-to-end тестирование
- [ ] Security audit
- [ ] Performance optimization
- [ ] Documentation

#### 4.2 Сообщество
- [ ] Public testnet
- [ ] Community nodes
- [ ] Documentation
- [ ] Support channels

### Фаза 5: Scaling (онлайн)
```
Month 14+: Community growth
```

#### 5.1 Масштабирование
- [ ] Federation support
- [ ] Bridge protocols
- [ ] Multi-network support
- [ ] Advanced features

---

## 🛠️ Технический стек

### Backend (Go)
```go
// Core dependencies
github.com/libp2p/go-libp2p
github.com/ipfs/go-ipfs-api
github.com/pion/webrtc
golang.org/x/crypto

// Network
github.com/txthinking/socks5
github.com/marten-seemann/quic-go
github.com/eyedeekay/go-i2p
```

### Frontend
```json
{
  "react": "^18.0.0",
  "tauri": "^1.0.0",
  "webrtc": "^4.0.0",
  "@react-native": "^0.70.0"
}
```

---

## 📊 Метрики успеха

### Технические метрики
- **Latency** < 100ms для локальных сообщений
- **Throughput** > 10MB/s для файлов
- **Uptime** > 99.9% для core nodes
- **Security** - zero known vulnerabilities

### Пользовательские метрики
- **Active users** - 1000+ в первый год
- **Messages/day** - 1M+ сообщений
- **Storage** - 1TB+ распределенного хранилища
- **Community nodes** - 100+ активных узлов

---

## 🔄 Риски и mitigation

### Технические риски
1. **NAT traversal complexity**
   - Mitigation: Multi-protocol fallback (STUN/TURN/I2P)

2. **IPFS performance**
   - Mitigation: Hybrid storage (IPFS + direct transfer)

3. **Crypto implementation**
   - Mitigation: External audit + formal verification

### Операционные риски
1. **Community adoption**
   - Mitigation: Open source + clear documentation

2. **Legal challenges**
   - Mitigation: Compliance with local laws

3. **Funding**
   - Mitigation: Community donations + grants

---

## 📅 Timeline

```
2024 Q2: MVP release
2024 Q3: Core features complete
2024 Q4: Mobile + Advanced protocols
2025 Q1: Production ready
2025 Q2+: Community growth
```

---

## 🤝 Contribution guidelines

### Для разработчиков
1. **Fork** проект на GitHub
2. **Create** feature branch
3. **Write** tests for new functionality
4. **Submit** pull request
5. **Code review** process

### Для сообщества
1. **Run** community nodes
2. **Provide** storage/relay services
3. **Report** bugs and issues
4. **Translate** documentation
5. **Spread** the word

---

*Этот roadmap является живым документом и может быть обновлен на основе feedback сообщества и технических требований.*
