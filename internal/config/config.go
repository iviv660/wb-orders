package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Kafka    KafkaConfig
}

type HTTPConfig struct {
	Addr string
}

type PostgresConfig struct {
	DSN string
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

// Load заполняет конфиг из переменных окружения с разумными значениями по умолчанию.
func Load() Config {
	return Config{
		HTTP: HTTPConfig{
			Addr: getenv("HTTP_ADDR", ":8080"),
		},
		Postgres: PostgresConfig{
			DSN: getenv("POSTGRES_DSN", "postgres://user:password@localhost:5432/wb?sslmode=disable"),
		},
		Kafka: KafkaConfig{
			Brokers: splitAndTrim(getenv("KAFKA_BROKERS", "localhost:9092")),
			Topic:   getenv("KAFKA_TOPIC", "orders"),
			GroupID: getenv("KAFKA_GROUP_ID", "orders-consumer"),
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
