package main

import (
	"log/slog"
	"os"
)

func main() {
	app := NewAppServer()

	port := os.Getenv("USER_SERVICE_PORT")
	slog.Info("Starting HTTP server on", "port:", port)
	app.RunHttpServer(port)
}
