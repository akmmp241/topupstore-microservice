package main

import (
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"os"
)

type AppServer struct {
	server   *fiber.App
	consumer *KafkaConsumer
}

func NewAppServer() *AppServer {
	server := fiber.New()

	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")
	bootstrapServer := kafkaHost + ":" + kafkaPort

	groupId := os.Getenv("USER_SERVICE_KAFKA_GROUP_ID")
	topic := os.Getenv("USER_SERVICE_KAFKA_TOPIC")

	consumer := NewKafkaConsumer(bootstrapServer, groupId, topic)

	return &AppServer{
		server:   server,
		consumer: consumer,
	}
}

func (a *AppServer) RunHttpServer(port string) {
	if err := a.server.Listen(":" + port); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func (a *AppServer) RunConsumer() {
	a.consumer.Read()
}
