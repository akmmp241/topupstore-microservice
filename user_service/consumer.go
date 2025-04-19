package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"io"
	"log/slog"
	"time"
)

type KafkaConsumer struct {
	Reader *kafka.Reader
}

func NewKafkaConsumer(bootstrapServer string, groupId string, topic string) *KafkaConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{bootstrapServer},
		Topic:          topic,
		GroupID:        groupId,
		CommitInterval: 1 * time.Second, // auto-commit interval
	})
	slog.Info("Kafka consumer created with", "topic:", topic, "group-id:", groupId)

	return &KafkaConsumer{
		Reader: r,
	}
}

func (c *KafkaConsumer) Read() {
	defer c.Reader.Close()
	slog.Info("Waiting for messages...")
	for {
		message, err := c.Reader.ReadMessage(context.Background())
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Warn("Reached EOF, possibly no messages yet.")
				continue
			}
			slog.Error("Error while reading", "error:", err)
			break
		}
		fmt.Println(string(message.Value))
	}
}
