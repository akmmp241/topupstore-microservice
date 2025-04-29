package shared

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"os"
	"time"
)

func getBootstrapServer() string {
	host := os.Getenv("KAFKA_HOST")
	port := os.Getenv("KAFKA_PORT")
	return fmt.Sprintf("%s:%s", host, port)
}

func NewKafkaConsumer(groupId string, topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{getBootstrapServer()},
		GroupID:        groupId,
		Topic:          topic,
		CommitInterval: 1 * time.Second,
	})
}

func NewProducer() *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(getBootstrapServer()),
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    10,
		BatchTimeout: time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}
}
