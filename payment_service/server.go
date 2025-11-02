package main

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
)

type AppServer struct {
	server *fiber.App
	db     *sql.DB
}

func NewAppServer(db *sql.DB) *AppServer {
	server := fiber.New(fiber.Config{
		ErrorHandler: shared.ErrorHandler,
	})
	validate := validator.New()

	api := server.Group("/api")

	paymentService := NewPaymentService(validate)
	paymentService.RegisterRoutes(api)

	return &AppServer{
		server: server,
		db:     db,
	}
}

func (a *AppServer) RunHttpServer(port string) {
	if err := a.server.Listen(":" + port); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
