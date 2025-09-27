# Heroin Android Client

Android клиент для secure P2P файлового хранилища и мессенджера.

## Архитектура

Модульная архитектура с чистым разделением ответственности:

### Core модули

- **core-crypto** - криптографические примитивы
  - XChaCha20-Poly1305 шифрование
  - X25519 key exchange
  - Ed25519 подписи
  - Argon2id KDF
  - Android Keystore интеграция

- **core-network** - P2P сеть и маршрутизация
  - libp2p клиент
  - QUIC/TCP/WSS транспорты
  - DHT discovery
  - Circuit relay v2
  - Адаптивная маршрутизация

- **core-storage** - IPFS интеграция
  - CAR файлы
  - Потоковое шифрование/дешифрование
  - Blake3 верификация
  - Локальный кэш

### Feature модули

- **feature-auth** - авторизация
  - Регистрация/логин
  - 2FA коды
  - Управление ключами
  
- **feature-messenger** - чаты
  - E2E сообщения
  - Double Ratchet
  - Групповые чаты
  
- **feature-files** - файлы
  - Загрузка/скачивание
  - Предпросмотр
  - Шеринг
  
- **feature-settings** - настройки
  - Сетевые настройки
  - Безопасность
  - Устройства

### App модуль

- Навигация (Compose Navigation)
- DI (Hilt)
- Темы (Material Design 3)
- Главная активность

## Технологии

- **Язык**: Kotlin
- **UI**: Jetpack Compose
- **Архитектура**: MVVM + Clean Architecture
- **Асинхронность**: Coroutines + Flow
- **DI**: Hilt
- **Сеть**: OkHttp + gRPC
- **Криптография**: libsodium + Android Keystore
- **P2P**: libp2p (через JNI или gRPC bridge)

## Сборка

### Требования

- Android Studio Flamingo+
- JDK 17
- Android SDK 34
- NDK (для crypto)

### Команды

```bash
# сборка debug
./gradlew assembleDebug

# сборка release
./gradlew assembleRelease

# тесты
./gradlew test

# установка на устройство
./gradlew installDebug
```

## Структура

```
client/android/
├── app/                    # Главный модуль
├── core-crypto/           # Криптография
├── core-network/          # P2P сеть
├── core-storage/          # IPFS хранилище
├── feature-auth/          # Авторизация
├── feature-messenger/     # Мессенджер
├── feature-files/         # Файлы
├── feature-settings/      # Настройки
└── build.gradle.kts       # Корневой build script
```

## Безопасность

### Хранение ключей

- Приватные ключи в Android Keystore
- Сессионные ключи в EncryptedSharedPreferences
- Резервные фразы только по запросу пользователя

### Сетевая безопасность

- Certificate pinning
- TLS 1.3 только
- Проверка подписей сообщений

### Защита от реверса

- ProGuard/R8 обфускация
- Native crypto через JNI
- Runtime Application Self-Protection

## UI/UX

### Дизайн

- Material Design 3
- Adaptive layouts
- Dark/Light themes
- Accessibility support

### Экраны

- Splash & Onboarding
- Auth (login/register/2FA)
- Chat list & Chat detail
- File browser & File detail
- Settings & Device management

## Разработка

### Добавление нового feature

1. Создать модуль `feature-name`
2. Добавить в `settings.gradle.kts`
3. Настроить DI модули
4. Добавить навигацию
5. Написать тесты

### Модульная структура

Каждый модуль содержит:
- `src/main/kotlin` - основной код
- `src/test/kotlin` - unit тесты
- `src/androidTest/kotlin` - инструментальные тесты
- `build.gradle.kts` - конфигурация сборки

## Тестирование

### Unit тесты

- Crypto функции
- Network логика
- ViewModels
- Use cases

### Integration тесты

- gRPC клиенты
- Database операции
- End-to-end flows

### UI тесты

- Compose тесты
- Espresso для сложных flow
- Screenshot тесты

## Релиз

### Подписание

- Release keystore в CI/CD
- Play App Signing
- Автоматические обновления

### Распространение

- Google Play Store
- F-Droid (open source build)
- Direct APK download
