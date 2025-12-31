package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig
	Postgres    PostgresConfig
	MongoDB     MongoDBConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Encryption  EncryptionConfig
	RateLimit   RateLimitConfig
	WebSocket   WebSocketConfig
	Moderation  ModerationConfig
	Timeouts    TimeoutConfig
}

type ServerConfig struct {
	Port         int
	Env          string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type TimeoutConfig struct {
	DB      time.Duration
	HTTP    time.Duration
	Context time.Duration
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

type MongoDBConfig struct {
	URI      string
	Database string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type EncryptionConfig struct {
	Key string
}

type RateLimitConfig struct {
	PostsPerHour     int
	ResponsesPerHour int
}

type WebSocketConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	MaxMessageSize  int
}

type ModerationConfig struct {
	EnableAutoModeration  bool
	ProfanityFilterLevel string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	accessExpiry, err := time.ParseDuration(viper.GetString("JWT_ACCESS_EXPIRY"))
	if err != nil {
		accessExpiry = 15 * time.Minute
	}

	refreshExpiry, err := time.ParseDuration(viper.GetString("JWT_REFRESH_EXPIRY"))
	if err != nil {
		refreshExpiry = 168 * time.Hour
	}

	readTimeout, _ := time.ParseDuration(viper.GetString("SERVER_READ_TIMEOUT"))
	writeTimeout, _ := time.ParseDuration(viper.GetString("SERVER_WRITE_TIMEOUT"))
	idleTimeout, _ := time.ParseDuration(viper.GetString("SERVER_IDLE_TIMEOUT"))
	dbTimeout, _ := time.ParseDuration(viper.GetString("DB_TIMEOUT"))
	httpTimeout, _ := time.ParseDuration(viper.GetString("HTTP_TIMEOUT"))
	contextTimeout, _ := time.ParseDuration(viper.GetString("CONTEXT_TIMEOUT"))

	cfg := &Config{
		Server: ServerConfig{
			Port:         viper.GetInt("SERVER_PORT"),
			Env:          viper.GetString("SERVER_ENV"),
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		Postgres: PostgresConfig{
			Host:     viper.GetString("POSTGRES_HOST"),
			Port:     viper.GetInt("POSTGRES_PORT"),
			User:     viper.GetString("POSTGRES_USER"),
			Password: viper.GetString("POSTGRES_PASSWORD"),
			Database: viper.GetString("POSTGRES_DB"),
			SSLMode:  viper.GetString("POSTGRES_SSL_MODE"),
		},
		MongoDB: MongoDBConfig{
			URI:      viper.GetString("MONGODB_URI"),
			Database: viper.GetString("MONGODB_DB"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetInt("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret:        viper.GetString("JWT_SECRET"),
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		Encryption: EncryptionConfig{
			Key: viper.GetString("ENCRYPTION_KEY"),
		},
		RateLimit: RateLimitConfig{
			PostsPerHour:     viper.GetInt("RATE_LIMIT_POSTS_PER_HOUR"),
			ResponsesPerHour: viper.GetInt("RATE_LIMIT_RESPONSES_PER_HOUR"),
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:  viper.GetInt("WS_READ_BUFFER_SIZE"),
			WriteBufferSize: viper.GetInt("WS_WRITE_BUFFER_SIZE"),
			MaxMessageSize:  viper.GetInt("WS_MAX_MESSAGE_SIZE"),
		},
		Moderation: ModerationConfig{
			EnableAutoModeration:  viper.GetBool("ENABLE_AUTO_MODERATION"),
			ProfanityFilterLevel: viper.GetString("PROFANITY_FILTER_LEVEL"),
		},
		Timeouts: TimeoutConfig{
			DB:      dbTimeout,
			HTTP:    httpTimeout,
			Context: contextTimeout,
		},
	}

	// Validate and apply defaults
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.Database, p.SSLMode)
}

// Validate performs comprehensive validation of all configuration settings
func (c *Config) Validate() error {
	// Server validation
	if c.Server.Port == 0 {
		return fmt.Errorf("SERVER_PORT is required")
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("SERVER_PORT must be between 1 and 65535")
	}
	if c.Server.Env == "" {
		c.Server.Env = "development"
	}
	if c.Server.Env != "development" && c.Server.Env != "staging" && c.Server.Env != "production" {
		return fmt.Errorf("SERVER_ENV must be one of: development, staging, production")
	}

	// Postgres validation
	if c.Postgres.Host == "" {
		return fmt.Errorf("POSTGRES_HOST is required")
	}
	if c.Postgres.Port == 0 {
		c.Postgres.Port = 5432
	}
	if c.Postgres.User == "" {
		return fmt.Errorf("POSTGRES_USER is required")
	}
	if c.Postgres.Password == "" {
		return fmt.Errorf("POSTGRES_PASSWORD is required")
	}
	if c.Postgres.Database == "" {
		return fmt.Errorf("POSTGRES_DB is required")
	}
	if c.Postgres.SSLMode == "" {
		c.Postgres.SSLMode = "disable"
	}

	// MongoDB validation
	if c.MongoDB.URI == "" {
		return fmt.Errorf("MONGODB_URI is required")
	}
	if c.MongoDB.Database == "" {
		return fmt.Errorf("MONGODB_DB is required")
	}

	// Redis validation
	if c.Redis.Host == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}
	if c.Redis.Port == 0 {
		c.Redis.Port = 6379
	}

	// JWT validation (critical security settings)
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required and must not be empty")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters for security")
	}
	if c.JWT.AccessExpiry == 0 {
		c.JWT.AccessExpiry = 15 * time.Minute
	}
	if c.JWT.RefreshExpiry == 0 {
		c.JWT.RefreshExpiry = 168 * time.Hour
	}
	if c.JWT.AccessExpiry > c.JWT.RefreshExpiry {
		return fmt.Errorf("JWT_ACCESS_EXPIRY cannot be greater than JWT_REFRESH_EXPIRY")
	}

	// Encryption validation (critical security settings)
	if c.Encryption.Key == "" {
		return fmt.Errorf("ENCRYPTION_KEY is required and must not be empty")
	}
	if len(c.Encryption.Key) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 bytes for AES-256")
	}

	// Rate limit defaults
	if c.RateLimit.PostsPerHour == 0 {
		c.RateLimit.PostsPerHour = 10
	}
	if c.RateLimit.ResponsesPerHour == 0 {
		c.RateLimit.ResponsesPerHour = 50
	}

	// WebSocket defaults
	if c.WebSocket.ReadBufferSize == 0 {
		c.WebSocket.ReadBufferSize = 1024
	}
	if c.WebSocket.WriteBufferSize == 0 {
		c.WebSocket.WriteBufferSize = 1024
	}
	if c.WebSocket.MaxMessageSize == 0 {
		c.WebSocket.MaxMessageSize = 8192
	}

	// Moderation defaults
	if c.Moderation.ProfanityFilterLevel == "" {
		c.Moderation.ProfanityFilterLevel = "medium"
	}

	// Server timeout defaults
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 15 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 15 * time.Second
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 60 * time.Second
	}

	// Timeout defaults
	if c.Timeouts.DB == 0 {
		c.Timeouts.DB = 10 * time.Second
	}
	if c.Timeouts.HTTP == 0 {
		c.Timeouts.HTTP = 30 * time.Second
	}
	if c.Timeouts.Context == 0 {
		c.Timeouts.Context = 30 * time.Second
	}

	return nil
}
