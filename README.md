# firehose-ingestor

A Go service that connects to the [Bluesky](https://bsky.app) AT Protocol firehose and streams new posts into a Kafka topic for downstream processing.

## How it works

The Bluesky firehose is a WebSocket stream of all public repo commits on the network. This service:

1. Connects to the firehose via WebSocket
1. Decodes incoming DAG-CBOR `.car` blocks using [indigo](https://github.com/bluesky-social/indigo)
1. Publishes each post to a Kafka topic

## Project structure

```
internal/
  firehose/   - WebSocket client and event runner
  kafka/      - Kafka producer
  models/     - Shared types (Event, Post)
```

## Prerequisites

- Go 1.21+
- A running Kafka broker

## Running Kafka locally

```bash
docker run -d --name broker \
  -p 9092:9092 \
  -e KAFKA_NODE_ID=1 \
  -e KAFKA_PROCESS_ROLES=broker,controller \
  -e KAFKA_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER \
  -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT \
  -e KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093 \
  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
  apache/kafka:latest
```

## Running the ingestor

```bash
go run ./cmd/ingestor
```

OR

```bash
make run
```

## Kafka message format

Messages are published to the `bluesky-posts` topic.

| Field | Value |
|-------|-------|
| Key   | The author's DID (e.g. `did:plc:abc123`) |
| Value | <subject to change> |

Using the DID as the key ensures all posts from the same user are routed to the same partition, preserving per-user ordering.

## Dependencies

- [indigo](https://github.com/bluesky-social/indigo) — AT Protocol client library
- [kafka-go](https://github.com/segmentio/kafka-go) — Kafka client
- [gorilla/websocket](https://github.com/gorilla/websocket) — WebSocket client