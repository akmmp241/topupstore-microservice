package shared

import (
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGrpcClientConn(target string) *grpc.ClientConn {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("error occurred while connect to grpc service", "err", err.Error())
		panic(err)
	}

	return conn
}
