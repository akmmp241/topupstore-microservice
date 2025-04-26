package main

import (
	"log/slog"
)

const PORT = "3002"

func main() {
	app := NewAppServer()

	slog.Info("Starting HTTP server on", "port:", PORT)
	app.RunHttpServer(PORT)
}
