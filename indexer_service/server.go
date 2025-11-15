package main

import (
	"os"
)

const GroupId = "indexer-service-group"

type AppServer struct {
	Consumer       *KafkaConsumer
	IndexerService *IndexerService
	EsClient       *ESClient
}

func NewAppServer(es *ESClient) *AppServer {
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")
	bootstrapServer := kafkaHost + ":" + kafkaPort

	return &AppServer{
		Consumer:       NewKafkaConsumer(bootstrapServer, GroupId),
		IndexerService: NewIndexerService(es),
	}
}

func (a *AppServer) RunProductIndexerConsumer() {
	a.Consumer.StartIndexerConsumer(a.IndexerService.HandleProductIndexer)
}
