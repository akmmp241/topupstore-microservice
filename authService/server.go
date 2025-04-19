package main

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"os"
)

type AppServer struct {
	db     *sql.DB
	server *fiber.App
}

func NewAppServer(producer *KafkaProducer) *AppServer {
	db := GetConnection()
	server := fiber.New()

	app := server.Group("/api/auth")

	authService := NewAuthService(producer)
	authService.RegisterRoutes(app)

	return &AppServer{
		db:     db,
		server: server,
	}
}

func (app *AppServer) Run(port string) {
	if err := app.server.Listen(":" + port); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
