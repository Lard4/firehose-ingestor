run:
	go run ./cmd/firehose

build:
	go build -o bin/firehose ./cmd/firehose

format:
	gofmt -w .