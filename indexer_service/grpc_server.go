package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"strconv"
	"time"

	ipb "github.com/akmmp241/topupstore-microservice/indexer-proto/v1"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	ListenAddr string
	Listener   net.Listener
	Server     *grpc.Server
	EsClient   *elasticsearch.Client
	ipb.UnimplementedIndexerServiceServer
}

func (g *GrpcServer) BulkIndexProducts(stream ipb.IndexerService_BulkIndexProductsServer) error {
	bulkIndexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:        g.EsClient,
		Index:         "products",
		NumWorkers:    4,
		FlushBytes:    1024 * 1024 * 5, // 5 MB
		FlushInterval: 30 * time.Second,
	})
	if err != nil {
		slog.Error("Error creating bulk indexer", "err", err)
		return err
	}

	var totalIndexed int64 = 0
	var totalFailed int64 = 0

	for {
		product, err := stream.Recv()
		if err == io.EOF {
			break // stream closed
		}

		if err != nil {
			slog.Warn("Error while reading from stream", "err", err)
			totalFailed++
			continue
		}

		data, err := json.Marshal(product)
		if err != nil {
			slog.Warn("Error marshalling product", "err", err)
			totalFailed++
			continue
		}

		err = bulkIndexer.Add(context.Background(), esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: strconv.FormatInt(int64(product.GetId()), 10),
			Body:       bytes.NewReader(data),
			OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
				totalIndexed++
			},
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				slog.Warn("Error indexing product", "err", err)
				totalFailed++
			},
		})
		if err != nil {
			slog.Warn("Error indexing product", "err", err)
			totalFailed++
		}
	}

	if err := bulkIndexer.Close(context.Background()); err != nil {
		slog.Error("Error closing bulk indexer", "err", err)
		return err
	}

	summary := ipb.BulkIndexSummary{
		TotalIndexed: uint64(totalIndexed),
		TotalFailed:  uint64(totalFailed),
	}

	slog.Info("Indexed products", "total_indexed", totalIndexed, "total_failed", totalFailed)
	return stream.SendAndClose(&summary)
}

func (g *GrpcServer) Start() {
	ipb.RegisterIndexerServiceServer(g.Server, g)

	slog.Info("Starting indexer service in gRPC server on port:", "port", g.ListenAddr)

	if err := g.Server.Serve(g.Listener); err != nil {
		slog.Error("Error occurred while serving gRPC server", "err", err)
		panic(err)
	}
}

func NewGrpcServer(listenAddr string) *GrpcServer {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		slog.Error("Error creating listener", "error", err)
		panic(err)
	}

	esClient := initElasticsearch()

	return &GrpcServer{
		ListenAddr: listenAddr,
		Listener:   listener,
		Server:     grpc.NewServer(),
		EsClient:   esClient,
	}
}
