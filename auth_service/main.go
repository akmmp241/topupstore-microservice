package main

import (
	"log/slog"
	"os"
)

const PORT = "3001"

func main() {
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")

	bootstrapServer := kafkaHost + ":" + kafkaPort

	slog.Info("Starting Kafka producer", "bootstrap-server", bootstrapServer)
	slog.Info("Starting HTTP server on port:", "port", PORT)
	producer := NewKafkaProducer(bootstrapServer)
	(NewAppServer(producer)).Run(PORT)
}
