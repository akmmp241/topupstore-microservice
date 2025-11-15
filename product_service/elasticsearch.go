package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

const ProductIndex = "products"

type ESClient struct {
	Client *elasticsearch.Client
}

func NewElasticsearch() *ESClient {
	esHost := os.Getenv("ES_HOST")
	esPort := os.Getenv("ES_PORT")

	cfg := elasticsearch.Config{
		Addresses: []string{fmt.Sprintf("http://%s:%s", esHost, esPort)},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		slog.Error("Error creating Elasticsearch client", "err", err)
		panic(err)
	}

	res, err := client.Info()
	if err != nil {
		slog.Error("Error getting response from Elasticsearch", "err", err)
		panic(err)
	}
	defer res.Body.Close()

	return &ESClient{
		Client: client,
	}
}
