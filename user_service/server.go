package main

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
)

type AppServer struct {
	server *fiber.App
	db     *sql.DB
}

func NewAppServer(db *sql.DB) *AppServer {
	server := fiber.New()
	validate := validator.New()

	api := server.Group("/api")

	userService := NewUserService(validate, db)
	userService.RegisterRoutes(api)

	return &AppServer{
		server: server,
	}
}

func (a *AppServer) RunHttpServer(port string) {
	slog.Info("Starting HTTP server on", "port:", port)
	if err := a.server.Listen(":" + port); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
