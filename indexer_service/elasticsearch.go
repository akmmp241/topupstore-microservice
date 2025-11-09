package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

//go:embed products.json
var productIndexMapping string

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

	es := &ESClient{
		Client: client,
	}
	es.RunMigrations()

	return es
}

func (c *ESClient) RunMigrations() {
	c.createIndexProduct()
}

func (c *ESClient) createIndexProduct() {
	response, err := c.Client.Indices.Exists([]string{ProductIndex})
	if err != nil {
		slog.Error("Error checking if index exists", "err", err)
		panic(err)
	}

	if response.StatusCode == 200 {
		slog.Info("Index already exists", "index", ProductIndex)
		return
	}

	if response.StatusCode == 404 {
		res, err := c.Client.Indices.Create(
			ProductIndex,
			c.Client.Indices.Create.WithBody(strings.NewReader(productIndexMapping)),
		)
		if err != nil {
			slog.Error("Error creating index", "err", err)
			panic(err)
		}
		defer res.Body.Close()

		if res.IsError() {
			slog.Error("Error creating index", "err", res.String())
			panic(res.String())
		}

		slog.Info("Index created", "index", ProductIndex)
		return
	}
}
