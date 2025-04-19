package main

import (
	"log/slog"
	"sync"
)

const PORT = "3002"

func main() {
	app := NewAppServer()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		slog.Info("Starting Kafka consumer")
		app.RunConsumer()
		defer wg.Done()
	}()

	wg.Add(1)
	go func() {
		slog.Info("Starting HTTP server on", "port:", PORT)
		app.RunHttpServer(PORT)
		defer wg.Done()
	}()

	wg.Wait()
}
