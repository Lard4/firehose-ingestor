run-firehose:
	go run ./cmd/firehose

build-firehose:
	go build -o bin/firehose ./cmd/firehose

run-reader:
	go run ./cmd/reader

build-reader:
	go build -o bin/reader ./cmd/reader

format:
	gofmt -w .