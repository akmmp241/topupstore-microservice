package main

import (
	"log/slog"
	"os"
	"sync"
)

const GroupId = "email-service-group"

type AppServer struct {
	Consumer     *KafkaConsumer
	EmailService *EmailService
}

func NewAppServer() *AppServer {
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")
	bootstrapServer := kafkaHost + ":" + kafkaPort

	return &AppServer{
		Consumer:     NewKafkaConsumer(bootstrapServer, GroupId),
		EmailService: NewEmailService(NewMailer()),
	}
}

func (a *AppServer) RunConsumer(wg *sync.WaitGroup) {
	wg.Add(2)

	go func() {
		slog.Info("Starting Auth Mail Consumer")
		a.Consumer.StartAuthConsumer(a.EmailService.HandleAuth)
		defer wg.Done()
	}()

	go func() {
		slog.Info("Starting New Order Mail Consumer")
		a.Consumer.StartOrderConsumer(a.EmailService.HandleOrder)
		defer wg.Done()
	}()
}
