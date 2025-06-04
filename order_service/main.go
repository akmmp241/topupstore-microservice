package main

import (
	"log/slog"
	"os"
)

func main() {
	port := os.Getenv("ORDER_SERVICE_PORT")
	slog.Info("Order service http server running on ", "port", port)
	(NewAppServer()).RunHttpServer(port)
}
