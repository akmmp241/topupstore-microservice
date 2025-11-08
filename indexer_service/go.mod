module github.com/akmmp241/topupstore-microservice/indexer-service

go 1.25.3

require (
	github.com/akmmp241/topupstore-microservice/indexer-proto v1.0.0
	github.com/elastic/go-elasticsearch/v8 v8.19.0
	google.golang.org/grpc v1.76.0
)

require (
	github.com/elastic/elastic-transport-go/v8 v8.7.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/akmmp241/topupstore-microservice/indexer-proto => ../indexer_proto
