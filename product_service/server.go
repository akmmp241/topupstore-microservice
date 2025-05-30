package main

import (
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"os"
)

type AppServer struct {
	server *fiber.App
}

func NewAppServer() *AppServer {
	db := shared.GetConnection()
	validate := validator.New()

	server := fiber.New(fiber.Config{
		ErrorHandler: shared.ErrorHandler,
	})

	app := server.Group("/api")

	productService := NewProductService(validate, db)
	productService.RegisterRoutes(app)

	return &AppServer{
		server: server,
	}
}

func (app *AppServer) Run(port string) {
	if err := app.server.Listen(":" + port); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
