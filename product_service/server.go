package main

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AppServer struct {
	server *fiber.App
	db     *sql.DB
}

func NewAppServer(db *sql.DB) *AppServer {
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
	slog.Info("Starting Product Service in HTTP server on port:", "port", port)
	if err := app.server.Listen(":" + port); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
