package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/segmentio/kafka-go"
)

const (
	ProductCreated = "product-created"
	ProductUpdated = "product-updated"
	ProductDeleted = "product-deleted"
)

type IndexerService struct {
	ESClient *elasticsearch.Client
}

func NewIndexerService(esClient *ESClient) *IndexerService {
	return &IndexerService{ESClient: esClient.Client}
}

func (s *IndexerService) HandleProductIndexer(msg *kafka.Message) error {
	var base BaseEvent[Product]

	if err := json.Unmarshal(msg.Value, &base); err != nil {
		slog.Error("Error unmarshalling message", "error", err)
		return err
	}

	data, err := json.Marshal(base.Data)
	if err != nil {
		slog.Error("Error marshalling data", "error", err)
	}

	switch base.EventType {
	case ProductCreated, ProductUpdated:
		return s.indexProduct(base.Data, bytes.NewReader(data))
	case ProductDeleted:
		return s.deleteProduct(base.Data)
	default:
		slog.Warn("Unknown event type", "event-type", base.EventType)
	}

	return nil
}

func (s *IndexerService) indexProduct(product Product, body io.Reader) error {
	req := esapi.IndexRequest{
		Index:      ProductIndex,
		DocumentID: product.ID,
		Body:       body,
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), s.ESClient)
	if err != nil {
		slog.Error("Error getting response for indexing product", "error", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		slog.Error("Error indexing product", "error", res.String())
		return err
	}

	slog.Info("Product indexed successfully", "product-id", product.ID)
	return nil
}

func (s *IndexerService) deleteProduct(product Product) error {
	req := esapi.DeleteRequest{
		Index:      ProductIndex,
		DocumentID: product.ID,
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), s.ESClient)
	if err != nil {
		slog.Error("Error getting response for deleting product", "error", err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		slog.Warn("Product not found for deletion", "product-id", product.ID)
		return nil
	}

	if res.IsError() {
		slog.Error("Error deleting product", "error", res.String())
		return err
	}

	slog.Info("Product deleted successfully", "product-id", product.ID)
	return nil
}
