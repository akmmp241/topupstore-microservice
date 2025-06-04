package main

import (
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"os"
)

type AppServer struct {
	server *fiber.App
}

func NewAppServer() *AppServer {
	db := shared.GetConnection()
	server := fiber.New(fiber.Config{
		ErrorHandler: shared.ErrorHandler,
	})
	validate := validator.New()

	api := server.Group("/api")

	producer := NewKafkaProducer()

	orderService := NewOrderService(db, validate, producer)
	orderService.RegisterRoutes(api)

	return &AppServer{
		server: server,
	}
}

func (a *AppServer) RunHttpServer(port string) {
	if err := a.server.Listen(":" + port); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
