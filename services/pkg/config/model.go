package config

// Config is the root configuration structure shared across all services.
// service configuration,
// logging configuration,
// database configuration,
//  queue configuration
type Config struct {
	Service  ServiceConfig  `mapstructure:"service"`
	Log      LogConfig      `mapstructure:"log"`
	Database DatabaseConfig `mapstructure:"database"`
	Queue    QueueConfig    `mapstructure:"queue"`
}

// ServiceConfig holds the configuration for the service itself.
type ServiceConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

// LogConfig holds the logging configuration.
type LogConfig struct {
	Level string `mapstructure:"level"` // debug | info | warn | error
}

// DatabaseConfig holds the configuration for the database connection.
type DatabaseConfig struct {
	Provider string        `mapstructure:"provider"` // postgres | mysql | mongodb | dynamodb
	Host     string        `mapstructure:"host"`
	Port     int           `mapstructure:"port"`
	Name     string        `mapstructure:"name"`
	User     string        `mapstructure:"user"`
	Password string        `mapstructure:"password"` // override via DATABASE_PASSWORD env var
	SSLMode  string        `mapstructure:"ssl_mode"`
	Pool     PoolConfig    `mapstructure:"pool"`
	Timeout  TimeoutConfig `mapstructure:"timeout"`
}

// PoolConfig holds the configuration for the database connection pool.
type PoolConfig struct {
	MaxOpen     int `mapstructure:"max_open"`
	MaxIdle     int `mapstructure:"max_idle"`
	MaxLifetime int `mapstructure:"max_lifetime"` // seconds
}

// TimeoutConfig holds the configuration for database connection timeouts.
type TimeoutConfig struct {
	Connect int `mapstructure:"connect"` // seconds
	Query   int `mapstructure:"query"`   // seconds
}

// queueConfig holds the configuration for the message queue connection.
type QueueConfig struct {
	Provider string         `mapstructure:"provider"` // rabbitmq | kafka | sqs | nats
	Host     string         `mapstructure:"host"`
	Port     int            `mapstructure:"port"`
	User     string         `mapstructure:"user"`
	Password string         `mapstructure:"password"` // override via QUEUE_PASSWORD env var
	Exchange ExchangeConfig `mapstructure:"exchange"`
	Queue    QueueSettings  `mapstructure:"queue"`
	Retry    RetryConfig    `mapstructure:"retry"`
	DLQ      DLQConfig      `mapstructure:"dlq"`
}

// exchangeConfig holds the configuration for the message queue exchange.
type ExchangeConfig struct {
	Name    string `mapstructure:"name"`
	Type    string `mapstructure:"type"` // topic | direct | fanout
	Durable bool   `mapstructure:"durable"`
}

// QueueSettings holds the configuration for the message queue.
type QueueSettings struct {
	Name       string `mapstructure:"name"`
	RoutingKey string `mapstructure:"routing_key"`
	Durable    bool   `mapstructure:"durable"`
}

// RetryConfig holds the configuration for retrying failed message processing.
type RetryConfig struct {
	MaxAttempts int `mapstructure:"max_attempts"`
	Interval    int `mapstructure:"interval"` // seconds
}

// DlqConfig holds the configuration for the dead-letter queue (DLQ).
type DLQConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Exchange string `mapstructure:"exchange"`
	Queue    string `mapstructure:"queue"`
}
