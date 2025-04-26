package main

import (
	"os"
	"sync"
)

const GroupId = "email-service-group"

type AppServer struct {
	consumer *KafkaConsumer
}

func NewAppServer() *AppServer {
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")
	bootstrapServer := kafkaHost + ":" + kafkaPort

	consumer := NewKafkaConsumer(bootstrapServer, GroupId)

	return &AppServer{
		consumer: consumer,
	}
}

func (a *AppServer) RunConsumer(wg *sync.WaitGroup) {
	wg.Add(2)

	go func() {
		a.consumer.StartUserRegistrationConsumer()
		defer wg.Done()
	}()

	go func() {
		a.consumer.StartUserLoginConsumer()
		defer wg.Done()
	}()
}
