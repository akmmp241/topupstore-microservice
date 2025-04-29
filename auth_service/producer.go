package main

import (
	"context"
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"os"
	"time"
)

type KafkaProducer struct {
	Writer *kafka.Writer
}

func NewKafkaProducer(bootstrapServer string) *KafkaProducer {
	topic := os.Getenv("USER_SERVICE_KAFKA_TOPIC")

	w := shared.NewProducer()
	slog.Info("Kafka producer created with", "topic:", topic)

	return &KafkaProducer{
		Writer: w,
	}
}

func (k *KafkaProducer) Write(ctx context.Context, topic string, messages ...[2]string) error {

	var msgs []kafka.Message

	for _, message := range messages {
		msgs = append(msgs, kafka.Message{
			Key:   []byte(message[0]),
			Value: []byte(message[1]),
			Topic: topic,
		})
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := k.Writer.WriteMessages(ctx, msgs...)
	if err != nil {
		return err
	}

	return nil
}
