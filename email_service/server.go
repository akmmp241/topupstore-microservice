package main

import (
	"os"
	"sync"
)

const GroupId = "email-service-group"

type AppServer struct {
	consumer     *KafkaConsumer
	EmailService *EmailService
}

func NewAppServer() *AppServer {
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")
	bootstrapServer := kafkaHost + ":" + kafkaPort

	return &AppServer{
		consumer:     NewKafkaConsumer(bootstrapServer, GroupId),
		EmailService: NewEmailService(NewMailer()),
	}
}

func (a *AppServer) RunConsumer(wg *sync.WaitGroup) {
	wg.Add(2)

	go func() {
		a.consumer.StartUserRegistrationConsumer(a.EmailService.HandleUserRegistration)
		defer wg.Done()
	}()

	go func() {
		a.consumer.StartUserLoginConsumer(a.EmailService.HandleUserLogin)
		defer wg.Done()
	}()
}
