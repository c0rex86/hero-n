# Contributing to HERO!N

🎉 Спасибо за интерес к проекту HERO!N! Мы рады любому вкладу в развитие по-настоящему свободной коммуникационной системы.

## 📋 Быстрый старт

### 1. Подготовка среды разработки

```bash
# Клонируйте репозиторий
git clone https://github.com/c0rex86/hero-n.git
cd hero-n

# Создайте ветку для ваших изменений
git checkout -b feature/your-feature-name

# Установите зависимости для backend
cd backend
go mod download

# Установите зависимости для mobile (Android Studio)
cd ../mobile
./gradlew build
```

### 2. Стиль кода

**Go (Backend):**
```go
// Используйте gofmt для форматирования
gofmt -w .

// Проверяйте с помощью golint
golint ./...

// Тестируйте с помощью go test
go test ./...
```

**Kotlin (Android):**
```kotlin
// Используйте официальные гайдлайны Kotlin
// https://kotlinlang.org/docs/coding-conventions.html

// Проверяйте с помощью ktlint
./gradlew ktlintCheck
```

## 🐛 Как сообщить о баге

### Шаг 1: Проверьте существующие issues
Убедитесь, что баг еще не был reported в [GitHub Issues](https://github.com/c0rex86/hero-n/issues).

### Шаг 2: Создайте новый issue
Используйте шаблон для баг-репорта и предоставьте:
- **Описание бага** - что произошло
- **Шаги воспроизведения** - как воспроизвести
- **Ожидаемое поведение** - что должно было произойти
- **Фактическое поведение** - что произошло на самом деле
- **Скриншоты/логи** - если применимо
- **Информация о среде** - ОС, версия приложения, устройство

### Шаг 3: Помогите с исправлением
Если у вас есть возможность, предложите fix или создайте pull request.

## ✨ Как предложить новую функцию

### Шаг 1: Обсудите идею
Создайте issue с тегом `enhancement` и опишите:
- **Проблему** - что эта функция решает
- **Предлагаемое решение** - как это должно работать
- **Альтернативы** - другие способы решения проблемы
- **Влияние** - как это повлияет на существующий код

### Шаг 2: Получите одобрение
Дождитесь feedback от maintainers и согласуйте детали реализации.

### Шаг 3: Реализуйте
Следуйте гайдлайнам разработки и создайте pull request.

## 🚀 Как создать Pull Request

### Шаг 1: Fork и clone
```bash
# Fork репозиторий через GitHub интерфейс
# Затем клонируйте ваш fork
git clone https://github.com/your-username/hero-n.git
cd hero-n

# Добавьте upstream remote
git remote add upstream https://github.com/c0rex86/hero-n.git
```

### Шаг 2: Создайте feature branch
```bash
# Всегда создавайте ветку от main
git checkout main
git pull upstream main

# Создайте ветку с описательным именем
git checkout -b feature/add-encryption-improvement
```

### Шаг 3: Разрабатывайте
```bash
# Делайте частые коммиты с описательными сообщениями
git commit -m "feat: improve encryption key rotation"

# Регулярно синхронизируйтесь с upstream
git pull upstream main
```

### Шаг 4: Подготовьте PR
```bash
# Убедитесь что все тесты проходят
go test ./...
./gradlew test

# Обновите документацию если нужно
# Добавьте тесты для новых функций

# Push вашу ветку
git push origin feature/add-encryption-improvement
```

### Шаг 5: Создайте Pull Request
1. Перейдите на GitHub и нажмите "Compare & pull request"
2. Заполните описание PR:
   - **Что** - что было изменено
   - **Почему** - зачем нужны эти изменения
   - **Как тестировать** - как проверить изменения
3. Назначьте reviewers
4. Дождитесь code review и approvals

## 📝 Гайдлайны для кода

### Общие правила

#### Для всех языков:
- **Читаемость** - код должен быть понятен другим разработчикам
- **Комментарии** - объясняйте сложную логику
- **Тестирование** - пишите тесты для новых функций
- **Безопасность** - следуйте best practices для криптографии

#### Go (Backend):
```go
// ✅ Хорошо
func (s *Server) HandleMessage(ctx context.Context, msg *Message) error {
    // Валидация входных данных
    if err := s.validateMessage(msg); err != nil {
        return fmt.Errorf("invalid message: %w", err)
    }

    // Обработка сообщения
    result, err := s.processMessage(ctx, msg)
    if err != nil {
        return fmt.Errorf("failed to process message: %w", err)
    }

    return nil
}

// ❌ Плохо
func (s *Server) HandleMessage(ctx context.Context, msg *Message) error {
    return s.processMessage(ctx, msg) // Нет валидации, неясно что происходит
}
```

#### Kotlin (Android):
```kotlin
// ✅ Хорошо
class MessageHandler(
    private val cryptoEngine: CryptoEngine,
    private val storage: MessageStorage
) {
    suspend fun handleMessage(message: Message): Result<Unit> {
        return try {
            // Шифрование
            val encrypted = cryptoEngine.encrypt(message.content)

            // Сохранение
            storage.saveMessage(message.copy(content = encrypted))

            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}

// ❌ Плохо
class MessageHandler {
    fun handleMessage(message: Message) {
        // Все в одном методе, сложно тестировать
    }
}
```

### Безопасность

#### Криптография:
- **Никогда не используйте устаревшие алгоритмы** (MD5, SHA-1, DES)
- **Всегда проверяйте подписи** перед обработкой данных
- **Используйте CSPRNG** для генерации ключей
- **Храните ключи безопасно** - никогда не в коде

#### Ввод данных:
- **Валидируйте все входные данные** перед обработкой
- **Используйте prepared statements** для SQL
- **Экранируйте HTML** в веб-интерфейсах
- **Ограничьте размер** загружаемых файлов

### Тестирование

#### Unit тесты:
```go
func TestMessageEncryption(t *testing.T) {
    crypto := NewCryptoEngine()
    message := []byte("Hello, World!")

    encrypted, err := crypto.Encrypt(message)
    require.NoError(t, err)
    require.NotEqual(t, message, encrypted)

    decrypted, err := crypto.Decrypt(encrypted)
    require.NoError(t, err)
    require.Equal(t, message, decrypted)
}
```

#### Integration тесты:
```go
func TestP2PMessageExchange(t *testing.T) {
    // Создаем два тестовых узла
    node1 := createTestNode()
    node2 := createTestNode()

    // Устанавливаем соединение
    err := node1.Connect(node2.ID())
    require.NoError(t, err)

    // Отправляем сообщение
    message := &Message{Content: "Test"}
    err = node1.SendMessage(node2.ID(), message)
    require.NoError(t, err)

    // Проверяем получение
    received, err := node2.ReceiveMessage()
    require.NoError(t, err)
    require.Equal(t, message.Content, received.Content)
}
```

## 🎨 Дизайн и UX

### Принципы дизайна
- **Простота** - интуитивно понятный интерфейс
- **Безопасность** - пользователь всегда знает о статусе шифрования
- **Надежность** - четкая индикация состояния сети
- **Доступность** - работает на различных устройствах

### Цветовая схема
- **Основной цвет** - темный с акцентами синего
- **Статус онлайн** - зеленый
- **Статус офлайн** - серый
- **Предупреждения** - оранжевый
- **Ошибки** - красный

## 📚 Документация

### Обновление документации
- **Всегда обновляйте** README при изменении API
- **Пишите примеры** использования новых функций
- **Обновляйте архитектуру** при значительных изменениях
- **Документируйте breaking changes** в CHANGELOG

### Стиль документации
```markdown
# Заголовок уровня 1

## Заголовок уровня 2

### Заголовок уровня 3

**Жирный текст** для важных моментов
*Курсив* для акцентов
`код` для команд и названий функций

> Цитата для важных замечаний

- Списки для перечислений
- Структурированная информация
- Легко читаемая
```

## 🚨 Безопасность - Сообщение о уязвимостях

Если вы обнаружили уязвимость безопасности:
1. **НЕ создавайте публичный issue**
2. **Отправьте email** на security@hero-n.dev
3. **Зашифруйте сообщение** используя наш PGP ключ
4. **Дождитесь подтверждения** перед публикацией

## 👥 Code of Conduct

Мы следуем [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).

### Наши ожидания:
- **Уважение** ко всем участникам
- **Конструктивная критика** вместо личных нападок
- **Терпимость** к разным точкам зрения
- **Профессионализм** в коммуникации

### Запрещено:
- Дискриминация по любому признаку
- Харрасмент и домогательства
- Публикация личной информации
- Спам и оффтоп

## 🏆 Recognition

### Contributors
Мы ценим вклад каждого участника и отмечаем:
- **Bug reporters** - находят и документируют проблемы
- **Code contributors** - пишут и улучшают код
- **Documentation writers** - создают и улучшают документацию
- **Testers** - тестируют и находят edge cases
- **Designers** - работают над UX/UI
- **Mentors** - помогают новым участникам

### Hall of Fame
Особо отличившиеся contributors получают:
- Упоминание в README
- Приглашение в core team
- Специальные бейджи на GitHub
- Приоритет при рассмотрении PR

## 📞 Контакты

- **GitHub Issues** - для багов и фич-реквестов
- **GitHub Discussions** - для обсуждений и вопросов
- **Matrix** - #hero-n:matrix.org для real-time общения
- **Email** - team@xder.c0rex64.dev для серьезных вопросов

## 🙏 Спасибо!

HERO!N - это сообщество людей, которые верят в свободную коммуникацию. Каждый вклад, будь то строка кода, баг-репорт или хорошая идея, приближает нас к цели создания по-настоящему устойчивой системы.

**Вместе мы создадим систему, которую нельзя сломать!**

c0re & команда x.d.e.r
