package main

import (
	"context"
	"errors"
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/segmentio/kafka-go"
	"io"
	"log/slog"
)

type HandlerKafka func(msg *kafka.Message) error

type KafkaConsumer struct {
	UserRegistrationReader *kafka.Reader
	UserLoginReader        *kafka.Reader
	ForgotPasswordReader   *kafka.Reader
	NewOrderReader         *kafka.Reader
	SuccessfulOrder        *kafka.Reader
	FailedOrder            *kafka.Reader
	EmailService           *EmailService
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
		UserRegistrationReader: initReader(kafkaConfig, "user-registration"),
		UserLoginReader:        initReader(kafkaConfig, "user-login"),
		ForgotPasswordReader:   initReader(kafkaConfig, "forget-password"),
		NewOrderReader:         initReader(kafkaConfig, "new_order"),
		SuccessfulOrder:        initReader(kafkaConfig, "order_succeeded"),
		FailedOrder:            initReader(kafkaConfig, "order_failed"),
	}
}

func initReader(cfg *KafkaConfig, topic string) *kafka.Reader {
	defer slog.Info("Kafka Consumer created with", "topic:", topic, "group-id:", cfg.GroupId)
	return shared.NewKafkaConsumer(cfg.GroupId, topic)
}

func (c *KafkaConsumer) StartUserRegistrationConsumer(handler HandlerKafka) {
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

		err = handler(&message)
		if err != nil {
			slog.Error("Error while handling message", "error:", err)
			continue
		}

		slog.Debug("Received message", "message:", string(message.Value), "key:", string(message.Key))
	}
}

func (c *KafkaConsumer) StartUserLoginConsumer(handler HandlerKafka) {
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

		err = handler(&message)
		if err != nil {
			slog.Error("Error while handling message", "error:", err)
			continue
		}

		slog.Debug("Received message", "message:", string(message.Value), "key", string(message.Key))

	}
}

func (c *KafkaConsumer) StartForgotPasswordConsumer(handler HandlerKafka) {
	defer c.ForgotPasswordReader.Close()
	for {
		message, err := c.ForgotPasswordReader.ReadMessage(context.Background())
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

func (c *KafkaConsumer) StartNewOrderConsumer(handler HandlerKafka) {
	defer c.NewOrderReader.Close()
	for {
		message, err := c.NewOrderReader.ReadMessage(context.Background())
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

func (c *KafkaConsumer) StartSuccessfulOrderConsumer(handler HandlerKafka) {
	defer c.SuccessfulOrder.Close()
	for {
		message, err := c.SuccessfulOrder.ReadMessage(context.Background())
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

func (c *KafkaConsumer) StartFailedOrderConsumer(handler HandlerKafka) {
	defer c.FailedOrder.Close()
	for {
		message, err := c.FailedOrder.ReadMessage(context.Background())
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
