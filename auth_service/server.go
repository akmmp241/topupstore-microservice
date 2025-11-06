package main

import (
	"log/slog"
	"os"

	"github.com/akmmp241/topupstore-microservice/shared"
	upb "github.com/akmmp241/topupstore-microservice/user-proto/v1"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AppServer struct {
	server *fiber.App
}

func NewAppServer() *AppServer {
	validate := validator.New()

	server := fiber.New(fiber.Config{
		ErrorHandler: shared.ErrorHandler,
	})

	app := server.Group("/api/auth")

	redisClient := shared.NewRedis()

	userServiceGrpcHost := os.Getenv("USER_SERVICE_GRPC_HOST")
	userServiceGrpcPort := os.Getenv("USER_SERVICE_GRPC_PORT")
	target := userServiceGrpcHost + ":" + userServiceGrpcPort
	conn := shared.NewGrpcClientConn(target)

	userServiceGrpc := upb.NewUserServiceClient(conn)

	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")
	bootstrapServer := kafkaHost + ":" + kafkaPort

	producer := NewKafkaProducer(bootstrapServer)

	authService := NewAuthService(producer, validate, redisClient, &userServiceGrpc)
	authService.RegisterRoutes(app)

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
