package main

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/segmentio/kafka-go"
)

const ProductIndexer = "product-indexer"

type HandlerKafka func(msg *kafka.Message) error

type KafkaConsumer struct {
	ProductIndexerReader *kafka.Reader
	IndexerService       *IndexerService
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
		ProductIndexerReader: initReader(kafkaConfig, ProductIndexer),
	}
}

func initReader(cfg *KafkaConfig, topic string) *kafka.Reader {
	defer slog.Info("Kafka Consumer created with", "topic:", topic, "group-id:", cfg.GroupId)
	return shared.NewKafkaConsumer(cfg.GroupId, topic)
}

func (c *KafkaConsumer) StartIndexerConsumer(handler HandlerKafka) {
	defer c.ProductIndexerReader.Close()
	for {
		message, err := c.ProductIndexerReader.ReadMessage(context.Background())
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Warn("Reached EOF, possibly no messages yet.")
				continue
			}
			slog.Error("Error reading message", "error", err)
			break
		}

		err = handler(&message)
		if err != nil {
			slog.Error("Error while handling message", "error", err)
			continue
		}

		slog.Debug("Received message", "message", string(message.Value), "key", string(message.Key))
	}
}

func (c *KafkaConsumer) Stop() {
	c.ProductIndexerReader.Close()
}
