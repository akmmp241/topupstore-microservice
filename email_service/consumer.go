package main

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/segmentio/kafka-go"
)

const (
	AuthTopic  = "auth-mail-service"
	OrderTopic = "order-mail-service"
)

type HandlerKafka func(msg *kafka.Message) error

type KafkaConsumer struct {
	AuthReader   *kafka.Reader
	OrderReader  *kafka.Reader
	EmailService *EmailService
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
		AuthReader:  initReader(kafkaConfig, AuthTopic),
		OrderReader: initReader(kafkaConfig, OrderTopic),
	}
}

func initReader(cfg *KafkaConfig, topic string) *kafka.Reader {
	defer slog.Info("Kafka Consumer created with", "topic:", topic, "group-id:", cfg.GroupId)
	return shared.NewKafkaConsumer(cfg.GroupId, topic)
}

func (c *KafkaConsumer) StartAuthConsumer(handler HandlerKafka) {
	defer c.AuthReader.Close()
	for {
		message, err := c.AuthReader.ReadMessage(context.Background())
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Warn("Reached EOF, possibly no messages yet.")
				continue
			}
			slog.Error("Error while reading", "error:", err)
			break
		}

		err = handler(&message)
		if err != nil {
			slog.Error("Error while handling message", "error:", err)
			continue
		}

		slog.Debug("Received message", "message:", string(message.Value), "key", string(message.Key))
	}
}

func (c *KafkaConsumer) StartOrderConsumer(handler HandlerKafka) {
	defer c.OrderReader.Close()
	for {
		message, err := c.OrderReader.ReadMessage(context.Background())
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Warn("Reached EOF, possibly no messages yet.")
				continue
			}
			slog.Error("Error while reading", "error:", err)
			break
		}

		err = handler(&message)
		if err != nil {
			slog.Error("Error while handling message", "error:", err)
			continue
		}

		slog.Debug("Received message", "message:", string(message.Value), "key", string(message.Key))
	}
}
