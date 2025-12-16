package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Env       string
	Inventory InventoryConfig
	Payment   PaymentConfig
	Postgres  PostgresConfig
	HTTP      HTTPConfig
	Logger    LoggerConfig
	Kafka     KafkaConfig
	Cache     CacheConfig
}

type InventoryConfig struct{}
type PaymentConfig struct{}

type PostgresConfig struct {
	DSN string
}

type HTTPConfig struct {
	Addr string
}

type LoggerConfig struct {
	Level  string
	AsJSON bool
}

type KafkaConfig struct {
	Brokers  []string
	Topic    string
	GroupID  string
	DLQTopic string
}

type CacheConfig struct {
	TTL time.Duration
}

var (
	once      sync.Once
	initErr   error
	AppConfig *Config
)

func Init() error {
	once.Do(func() {
		c := load()
		AppConfig = &c
	})
	return initErr
}

func MustInit() {
	if err := Init(); err != nil {
		panic(err)
	}
}

func Get() *Config {
	_ = Init()
	return AppConfig
}

func load() Config {
	return Config{
		Env: getenv("APP_ENV", "local"),
		Postgres: PostgresConfig{
			DSN: getenv("POSTGRES_DSN", "postgres://user:password@localhost:5432/wb?sslmode=disable"),
		},
		HTTP: HTTPConfig{
			Addr: getenv("HTTP_ADDR", ":8080"),
		},
		Logger: LoggerConfig{
			Level:  getenv("LOG_LEVEL", "info"),
			AsJSON: getbool("LOG_JSON", false),
		},
		Kafka: KafkaConfig{
			Brokers:  splitAndTrim(getenv("KAFKA_BROKERS", "localhost:9092")),
			Topic:    getenv("KAFKA_TOPIC", "orders"),
			GroupID:  getenv("KAFKA_GROUP_ID", "orders-consumer"),
			DLQTopic: getenv("KAFKA_DLQ_TOPIC", "orders.dlq"),
		},
		Cache: CacheConfig{
			TTL: getduration("CACHE_TTL", 5*time.Minute),
		},
	}
}

func getenv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func getbool(key string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return def
	}
	if v == "true" || v == "1" || v == "yes" || v == "y" || v == "on" {
		return true
	}
	if v == "false" || v == "0" || v == "no" || v == "n" || v == "off" {
		return false
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getduration(key string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return def
	}
	return d
}

func splitAndTrim(val string) []string {
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
