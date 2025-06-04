package main

import (
	"log/slog"
	"os"
)

func main() {
	port := os.Getenv("PAYMENT_SERVICE_PORT")
	slog.Info("Payment service http server running on ", "port", port)
	(NewAppServer()).RunHttpServer(port)
}
