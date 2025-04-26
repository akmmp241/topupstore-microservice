package main

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"io"
	"log/slog"
	"time"
)

type KafkaConsumer struct {
	UserRegistrationReader *kafka.Reader
	UserLoginReader        *kafka.Reader
}

type KafkaConfig struct {
	BootstrapServer string
	GroupId         string
}

func NewKafkaConsumer(bootstrapServer string, groupId string) *KafkaConsumer {
	kafkaConfig := &KafkaConfig{
		BootstrapServer: bootstrapServer,
		GroupId:         groupId,
	}

	return &KafkaConsumer{
		UserRegistrationReader: initUserRegistrationReader(kafkaConfig),
		UserLoginReader:        initUserLoginReader(kafkaConfig),
	}
}

func initUserRegistrationReader(cfg *KafkaConfig) *kafka.Reader {
	defer slog.Info("Kafka consumer created with", "topic:", "user-registration", "group-id:", cfg.GroupId)
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cfg.BootstrapServer},
		GroupID:        cfg.GroupId,
		Topic:          "user-registration",
		CommitInterval: 1 * time.Second,
	})
}

func initUserLoginReader(cfg *KafkaConfig) *kafka.Reader {
	defer slog.Info("Kafka consumer created with", "topic:", "user-login", "group-id:", cfg.GroupId)
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cfg.BootstrapServer},
		GroupID:        cfg.GroupId,
		Topic:          "user-login",
		CommitInterval: 1 * time.Second,
	})
}

func (c *KafkaConsumer) StartUserRegistrationConsumer() {
	defer c.UserRegistrationReader.Close()
	for {
		message, err := c.UserRegistrationReader.ReadMessage(context.Background())
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Warn("Reached EOF, possibly no messages yet.")
				continue
			}
			slog.Error("Error while reading", "error:", err)
			break
		}
		slog.Debug("Received message", "message:", string(message.Value), "key:", string(message.Key))
	}
}

func (c *KafkaConsumer) StartUserLoginConsumer() {
	defer c.UserLoginReader.Close()
	for {
		message, err := c.UserLoginReader.ReadMessage(context.Background())
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Warn("Reached EOF, possibly no messages yet.")
				continue
			}
			slog.Error("Error while reading", "error:", err)
			break
		}
		slog.Debug("Received message", "message:", string(message.Value), "key", string(message.Key))
	}
}
