package models

import (
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaMessage wraps the built-in Kafka message and adds any
// necessary additional fields.
type KafkaMessage struct {
	kafka.Message
	RowID        string
	ReceivedTime time.Time
}

func (km KafkaMessage) Latency() time.Duration {
	return km.ReceivedTime.Sub(km.Time)
}
