package configs

import (
	"os"
	"strconv"
)

// Config - основная конфигурация приложения
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	P2P      P2PConfig
	Crypto   CryptoConfig
	Logging  LoggingConfig
}

// ServerConfig - конфигурация HTTP сервера
type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  int // секунды
	WriteTimeout int // секунды
}

// DatabaseConfig - конфигурация базы данных
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// P2PConfig - конфигурация P2P сети
type P2PConfig struct {
	Port            string
	BootstrapPeers  []string
	DiscoveryEnabled bool
	RendezvousPoint  string
}

// CryptoConfig - конфигурация криптографии
type CryptoConfig struct {
	KeySize       int
	HashAlgorithm string
}

// LoggingConfig - конфигурация логирования
type LoggingConfig struct {
	Level  string
	Format string
}

// LoadConfig - загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT", 15),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT", 15),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "hero_user"),
			Password: getEnv("DB_PASSWORD", "hero_pass"),
			Name:     getEnv("DB_NAME", "hero_n"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		P2P: P2PConfig{
			Port:            getEnv("P2P_PORT", "4001"),
			BootstrapPeers:  []string{}, // TODO: добавить bootstrap пиры
			DiscoveryEnabled: getEnvBool("P2P_DISCOVERY_ENABLED", true),
			RendezvousPoint:  getEnv("P2P_RENDEZVOUS", "/hero-n/1.0.0"),
		},
		Crypto: CryptoConfig{
			KeySize:       getEnvInt("CRYPTO_KEY_SIZE", 32),
			HashAlgorithm: getEnv("CRYPTO_HASH_ALGO", "sha256"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
	}
}

// getEnv - получает переменную окружения с дефолтным значением
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt - получает переменную окружения как int
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool - получает переменную окружения как bool
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
