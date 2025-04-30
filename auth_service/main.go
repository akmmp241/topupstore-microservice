package main

import (
	"log/slog"
	"os"
)

func main() {
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")

	bootstrapServer := kafkaHost + ":" + kafkaPort

	slog.Info("Starting Kafka producer", "bootstrap-server", bootstrapServer)
	producer := NewKafkaProducer(bootstrapServer)

	port := os.Getenv("AUTH_SERVICE_PORT")
	slog.Info("Starting HTTP server on port:", "port", port)
	(NewAppServer(producer)).Run(port)
}
