package main

import "os"

func main() {
	grpcPort := os.Getenv("INDEXER_SERVICE_GRPC_PORT")

	NewGrpcServer(":" + grpcPort).Start()
}
