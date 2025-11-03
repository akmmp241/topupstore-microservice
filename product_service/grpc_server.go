package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net"
	"time"

	prpb "github.com/akmmp241/topupstore-microservice/product-proto/v1"
	"github.com/akmmp241/topupstore-microservice/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcServer struct {
	ListenAddr string
	DB         *sql.DB
	Server     *grpc.Server
	Listener   net.Listener
	prpb.UnimplementedProductServiceServer
}

func (g *GrpcServer) GetProductById(ctx context.Context, req *prpb.GetProductByIdReq) (*prpb.GetProductByIdRes, error) {
	tx, err := g.DB.Begin()
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return nil, err
	}
	defer shared.CommitOrRollback(tx, err)

	query := "SELECT id, ref_id, product_type_id, name, description, image_url, price, created_at, updated_at FROM products WHERE id = ?"
	row := g.DB.QueryRowContext(ctx, query, req.GetProductId())

	var product prpb.Product
	var createdAt, updatedAt time.Time

	if err := row.Scan(&product.Id, &product.RefId, &product.ProductTypeId, &product.Name, &product.Description, &product.ImageUrl, &product.Price, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "Product not found")
		}
		slog.Error("Failed to scan product row", "error", err)
		return nil, err
	}
	product.CreatedAt = timestamppb.New(createdAt)
	product.UpdatedAt = timestamppb.New(updatedAt)

	return &prpb.GetProductByIdRes{Product: &product}, nil
}

func (g *GrpcServer) Run() {
	prpb.RegisterProductServiceServer(g.Server, g)

	slog.Info("Starting Product Service in gRPC server on port:", "port", g.ListenAddr)

	if err := g.Server.Serve(g.Listener); err != nil {
		slog.Error("Error occurred while serving gRPC server", "err", err)
		panic(err)
	}
}

func NewGrpcServer(addr string, DB *sql.DB) *GrpcServer {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("Error occurred while creating listener", "err", err)
		panic(err)
	}

	return &GrpcServer{
		ListenAddr: addr,
		DB:         DB,
		Server:     grpc.NewServer(),
		Listener:   listener,
	}
}
