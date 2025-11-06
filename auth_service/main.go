package main

import (
	"log/slog"
	"os"
)

func main() {
	port := os.Getenv("AUTH_SERVICE_PORT")
	slog.Info("Starting HTTP server on port:", "port", port)
	(NewAppServer()).Run(port)
}
