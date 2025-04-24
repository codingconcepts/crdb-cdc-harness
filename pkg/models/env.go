package models

import "time"

// EnvironmentVariables will be populated from the user's environment
// on launch.
type EnvironmentVariables struct {
	DatabaseURL        string        `env:"DATABASE_URL"`
	KafkaURL           string        `env:"KAFKA_URL"`
	KafkaTopic         string        `env:"KAFKA_TOPIC"`
	KafkaMessageIDPath string        `env:"KAFKA_MESSAGE_ID_PATH"`
	WriteFrequency     time.Duration `env:"WRITE_FREQUENCY"`
	LatencyPoolSize    int           `env:"LATENCY_POOL_SIZE"`
	WriteStatement     string        `env:"WRITE_STATEMENT"`
	Args               StringFlags   `env:"ARGS"`
}
