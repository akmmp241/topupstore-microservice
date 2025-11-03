package main

import (
	"log/slog"
	"os"

	ppb "github.com/akmmp241/topupstore-microservice/payment-proto/v1"
	prpb "github.com/akmmp241/topupstore-microservice/product-proto/v1"
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
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

	paymentServiceGrpcHost := os.Getenv("PAYMENT_SERVICE_GRPC_HOST")
	paymentServiceGrpcPort := os.Getenv("PAYMENT_SERVICE_GRPC_PORT")
	paymentTarget := paymentServiceGrpcHost + ":" + paymentServiceGrpcPort
	paymentConn := shared.NewGrpcClientConn(paymentTarget)

	paymentServiceGrpc := ppb.NewPaymentServiceClient(paymentConn)

	productServiceGrpcHost := os.Getenv("PRODUCT_SERVICE_GRPC_HOST")
	productServiceGrpcPort := os.Getenv("PRODUCT_SERVICE_GRPC_PORT")
	productTarget := productServiceGrpcHost + ":" + productServiceGrpcPort
	productConn := shared.NewGrpcClientConn(productTarget)

	productServiceGrpc := prpb.NewProductServiceClient(productConn)

	orderService := NewOrderService(db, validate, producer, &paymentServiceGrpc, &productServiceGrpc)
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
