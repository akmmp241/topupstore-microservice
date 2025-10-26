package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akmmp241/topupstore-microservice/shared"
)

func main() {
	port := os.Getenv("USER_SERVICE_PORT")
	grpcPort := os.Getenv("USER_SERVICE_GRPC_PORT")

	db := shared.GetConnection()

	app := NewAppServer(db)
	grpcApp := NewGrpcServer(":"+grpcPort, db)

	go app.RunHttpServer(port)
	go grpcApp.Run()

	gracefulShutdown(app, grpcApp, db)
}

func gracefulShutdown(app *AppServer, grpcApp *GrpcServer, db *sql.DB) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.server.ShutdownWithContext(ctx); err != nil {
		slog.Error("failed to gracefully shutdown the server", "err", err)
	}
	slog.Info("http gracefully shut down")

	{
		grpcApp.Server.GracefulStop()
		slog.Info("grpc gracefully shut down")
	}

	if err := db.Close(); err != nil {
		slog.Error("failed to close db connection", "err", err)
	}
	slog.Info("db connection closed")

	slog.Info("server stopped")
}
