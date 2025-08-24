package storage

// SQL скрипты для создания таблиц
const (
	CreateUsersTable = `
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(64) PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			public_key TEXT NOT NULL,
			display_name VARCHAR(100) NOT NULL,
			status VARCHAR(20) DEFAULT 'offline',
			last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Индексы для таблицы пользователей
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
		CREATE INDEX IF NOT EXISTS idx_users_last_seen ON users(last_seen);
	`

	CreateMessagesTable = `
		CREATE TABLE IF NOT EXISTS messages (
			id VARCHAR(64) PRIMARY KEY,
			from_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			to_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			encrypted_content TEXT,
			signature TEXT,
			type VARCHAR(20) DEFAULT 'text',
			status VARCHAR(20) DEFAULT 'sent',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL
		);

		-- Индексы для таблицы сообщений
		CREATE INDEX IF NOT EXISTS idx_messages_from_to ON messages(from_id, to_id);
		CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
		CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
		CREATE INDEX IF NOT EXISTS idx_messages_deleted_at ON messages(deleted_at);

		-- Составной индекс для быстрого поиска диалогов
		CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(
			LEAST(from_id, to_id),
			GREATEST(from_id, to_id),
			created_at DESC
		);
	`

	CreateUserSessionsTable = `
		CREATE TABLE IF NOT EXISTS user_sessions (
			id VARCHAR(64) PRIMARY KEY,
			user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token TEXT NOT NULL,
			ip_address INET,
			user_agent TEXT,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			UNIQUE(user_id, token)
		);

		-- Индексы для сессий
		CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON user_sessions(user_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON user_sessions(expires_at);
		CREATE INDEX IF NOT EXISTS idx_sessions_token ON user_sessions(token);
	`

	CreateContactsTable = `
		CREATE TABLE IF NOT EXISTS contacts (
			id VARCHAR(64) PRIMARY KEY,
			user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			contact_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			nickname VARCHAR(100),
			added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			blocked BOOLEAN DEFAULT FALSE,

			UNIQUE(user_id, contact_id)
		);

		-- Индексы для контактов
		CREATE INDEX IF NOT EXISTS idx_contacts_user_id ON contacts(user_id);
		CREATE INDEX IF NOT EXISTS idx_contacts_contact_id ON contacts(contact_id);
		CREATE INDEX IF NOT EXISTS idx_contacts_blocked ON contacts(blocked);
	`

	// Триггеры для автоматического обновления updated_at
	CreateTriggers = `
		-- Триггер для пользователей
		CREATE OR REPLACE FUNCTION update_users_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		CREATE TRIGGER trigger_users_updated_at
			BEFORE UPDATE ON users
			FOR EACH ROW
			EXECUTE FUNCTION update_users_updated_at();

		-- Триггер для сообщений
		CREATE OR REPLACE FUNCTION update_messages_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		CREATE TRIGGER trigger_messages_updated_at
			BEFORE UPDATE ON messages
			FOR EACH ROW
			EXECUTE FUNCTION update_messages_updated_at();
	`

	// Индексы для производительности
	CreatePerformanceIndexes = `
		-- Индекс для поиска последних сообщений в диалоге
		CREATE INDEX IF NOT EXISTS idx_messages_recent ON messages(
			GREATEST(from_id, to_id),
			LEAST(from_id, to_id),
			created_at DESC
		);

		-- Частичный индекс для непрочитанных сообщений
		CREATE INDEX IF NOT EXISTS idx_messages_unread ON messages(to_id, created_at)
		WHERE status = 'sent';

		-- Индекс для поиска по типу сообщения
		CREATE INDEX IF NOT EXISTS idx_messages_type ON messages(type, created_at DESC);
	`
)

// SQL запросы для работы с данными
const (
	// Вставка пользователя
	InsertUserQuery = `
		INSERT INTO users (id, username, public_key, display_name, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	// Получение пользователя по ID
	GetUserByIDQuery = `
		SELECT id, username, public_key, display_name, status, last_seen, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	// Получение пользователя по username
	GetUserByUsernameQuery = `
		SELECT id, username, public_key, display_name, status, last_seen, created_at, updated_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`

	// Обновление статуса пользователя
	UpdateUserStatusQuery = `
		UPDATE users
		SET status = $2, last_seen = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	// Вставка сообщения
	InsertMessageQuery = `
		INSERT INTO messages (id, from_id, to_id, content, encrypted_content, signature, type, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	// Получение сообщения по ID
	GetMessageByIDQuery = `
		SELECT id, from_id, to_id, content, encrypted_content, signature, type, status, created_at, updated_at
		FROM messages
		WHERE id = $1 AND deleted_at IS NULL
	`

	// Получение сообщений между двумя пользователями
	GetMessagesBetweenUsersQuery = `
		SELECT id, from_id, to_id, content, encrypted_content, signature, type, status, created_at, updated_at
		FROM messages
		WHERE ((from_id = $1 AND to_id = $2) OR (from_id = $2 AND to_id = $1))
		AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	// Получение последних сообщений пользователя
	GetRecentMessagesQuery = `
		SELECT DISTINCT ON (conversation_id) id, from_id, to_id, content, type, status, created_at
		FROM (
			SELECT *,
				CASE
					WHEN from_id < to_id THEN from_id || '_' || to_id
					ELSE to_id || '_' || from_id
				END as conversation_id
			FROM messages
			WHERE (from_id = $1 OR to_id = $1) AND deleted_at IS NULL
		) AS conversations
		ORDER BY conversation_id, created_at DESC
		LIMIT $2
	`
)
