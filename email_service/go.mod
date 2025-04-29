module github.com/akmmp241/topupstore-microservice/email-service

go 1.23.4

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
	github.com/akmmp241/topupstore-microservice/shared v1.0.0
)

replace (
	github.com/akmmp241/topupstore-microservice/shared => ../shared
)