package config

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// keep fields aligned with configs/config.example.yaml

type ServerListenConfig struct {
	TCP  string `yaml:"tcp"`
	QUIC string `yaml:"quic"`
}

type TLSServerConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type TransportsConfig struct {
	EnableTCP bool `yaml:"enable_tcp"`
	EnableQUIC bool `yaml:"enable_quic"`
	EnableWS  bool `yaml:"enable_ws"`
	EnableWSS bool `yaml:"enable_wss"`
}

type ServerConfig struct {
	Listen     ServerListenConfig `yaml:"listen"`
	TLS        TLSServerConfig    `yaml:"tls"`
	Transports TransportsConfig   `yaml:"transports"`
}

type RoutingConfig struct {
	Strategy           string   `yaml:"strategy"`
	BootstrapNodes     []string `yaml:"bootstrap_nodes"`
	Relays             []string `yaml:"relays"`
	SwitchThresholdMS  int      `yaml:"switch_threshold_ms"`
	MetricsWindowSec   int      `yaml:"metrics_window_sec"`
}

type IPFSConfig struct {
	Endpoint          string `yaml:"endpoint"`
	PinningEnabled    bool   `yaml:"pinning_enabled"`
	ReplicationFactor int    `yaml:"replication_factor"`
}

type KDFConfig struct {
	Type     string `yaml:"type"`
	Time     uint32 `yaml:"time"`
	MemoryMB uint32 `yaml:"memory_mb"`
	Threads  uint8  `yaml:"threads"`
	KeyLen   uint32 `yaml:"key_len"`
}

type TokenConfig struct {
	Issuer             string `yaml:"issuer"`
	LifetimeMin        int    `yaml:"lifetime_min"`
	RefreshDays        int    `yaml:"refresh_days"`
	SymmetricKeyBase64 string `yaml:"symmetric_key_base64"`
}

type SecondaryKeyConfig struct {
	Length              int `yaml:"length"`
	RotateMinutes       int `yaml:"rotate_minutes"`
	AllowedClockSkewSec int `yaml:"allowed_clock_skew_sec"`
}

type SecurityConfig struct {
	KDF            KDFConfig          `yaml:"kdf"`
	Token          TokenConfig        `yaml:"token"`
	SecondaryKey   SecondaryKeyConfig `yaml:"secondary_key"`
	TLSFingerprint string            `yaml:"tls_fingerprint"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type ObservabilityConfig struct {
	PrometheusAddr string `yaml:"prometheus_addr"`
	OTLPEndpoint   string `yaml:"otlp_endpoint"`
}

type Config struct {
	Server        ServerConfig       `yaml:"server"`
	Routing       RoutingConfig      `yaml:"routing"`
	IPFS          IPFSConfig         `yaml:"ipfs"`
	Security      SecurityConfig     `yaml:"security"`
	Database      DatabaseConfig     `yaml:"database"`
	Logging       LoggingConfig      `yaml:"logging"`
	Observability ObservabilityConfig `yaml:"observability"`

	// derived
	pasetoSymmetricKey []byte
	issuerEd25519Priv  ed25519.PrivateKey
	issuerEd25519Pub   ed25519.PublicKey
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	if err := c.populateDerived(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) populateDerived() error {
	if c.Security.Token.SymmetricKeyBase64 != "" {
		key, err := base64.StdEncoding.DecodeString(c.Security.Token.SymmetricKeyBase64)
		if err != nil {
			return fmt.Errorf("decode token symmetric key: %w", err)
		}
		if len(key) != 32 {
			return fmt.Errorf("token symmetric key must be 32 bytes after base64 decoding")
		}
		c.pasetoSymmetricKey = key
	} else {
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			return fmt.Errorf("generate token symmetric key: %w", err)
		}
		c.pasetoSymmetricKey = buf
	}
	if c.Security.Token.LifetimeMin <= 0 {
		c.Security.Token.LifetimeMin = 30
	}
	return nil
}

func (c *Config) AccessTokenTTL() time.Duration {
	return time.Duration(c.Security.Token.LifetimeMin) * time.Minute
}

func (c *Config) PasetoKey() []byte {
	return c.pasetoSymmetricKey
}
