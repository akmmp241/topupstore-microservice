module github.com/akmmp241/topupstore-microservice/product-service

go 1.25.3

require (
	github.com/akmmp241/topupstore-microservice/product-proto v1.0.0
	github.com/akmmp241/topupstore-microservice/shared v1.0.0
	github.com/go-playground/validator/v10 v10.26.0
	github.com/gofiber/fiber/v2 v2.52.8
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-sql-driver/mysql v1.9.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/redis/go-redis/v9 v9.9.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/segmentio/kafka-go v0.4.48 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.62.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/grpc v1.76.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace (
	github.com/akmmp241/topupstore-microservice/product-proto => ../product_proto
	github.com/akmmp241/topupstore-microservice/shared => ../shared
)
