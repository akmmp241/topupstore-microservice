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
		slog.Info("Starting User Registration Consumer")
		a.Consumer.StartUserRegistrationConsumer(a.EmailService.HandleUserRegistration)
		defer wg.Done()
	}()

	go func() {
		slog.Info("Starting User Login Consumer")
		a.Consumer.StartUserLoginConsumer(a.EmailService.HandleUserLogin)
		defer wg.Done()
	}()
}
