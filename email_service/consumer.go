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
		UserRegistrationReader: initUserRegistrationReader(kafkaConfig),
		UserLoginReader:        initUserLoginReader(kafkaConfig),
		ForgotPasswordReader:   initForgotPasswordReader(kafkaConfig),
		NewOrderReader:         initNewOrderReader(kafkaConfig),
	}
}

func initUserRegistrationReader(cfg *KafkaConfig) *kafka.Reader {
	defer slog.Info("Kafka Consumer created with", "topic:", "user-registration", "group-id:", cfg.GroupId)
	return shared.NewKafkaConsumer(cfg.GroupId, "user-registration")
}

func initUserLoginReader(cfg *KafkaConfig) *kafka.Reader {
	defer slog.Info("Kafka Consumer created with", "topic:", "user-login", "group-id:", cfg.GroupId)
	return shared.NewKafkaConsumer(cfg.GroupId, "user-login")
}

func initForgotPasswordReader(cfg *KafkaConfig) *kafka.Reader {
	defer slog.Info("Kafka Consumer created with", "topic:", "forgot-password", "group-id:", cfg.GroupId)
	return shared.NewKafkaConsumer(cfg.GroupId, "forget-password")
}

func initNewOrderReader(cfg *KafkaConfig) *kafka.Reader {
	defer slog.Info("Kafka Consumer created with", "topic:", "new-order", "group-id:", cfg.GroupId)
	return shared.NewKafkaConsumer(cfg.GroupId, "new_order")
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
