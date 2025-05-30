package main

import (
	"log/slog"
	"os"
)

func main() {
	port := os.Getenv("PRODUCT_SERVICE_PORT")
	slog.Info("Starting Product Service in HTTP server on port:", "port", port)
	(NewAppServer()).Run(port)
}
