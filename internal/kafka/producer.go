package kafka

import (
	"context"
	"time"

	"github.com/lard4/firehose-ingestor/internal/models"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer() *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:  kafka.TCP("localhost:9092"),
			Topic: "bluesky-posts",
		},
	}
}

func (p *KafkaProducer) Send(ctx context.Context, event models.Event) error {
	postTime, err := time.Parse(time.RFC3339, event.Post.CreatedAt)
	if err != nil {
		postTime = time.Now() // fallback
	}

	return p.writer.WriteMessages(
		ctx,
		kafka.Message{
			// Important: We use the DID as the key so that all posts from the same user go to the same partition, which keeps the ordering of their posts consistent.
			Key: []byte(event.DID),
			// For now, the value is just the post text, but we may want to serialize since it could be an image.
			Value: []byte(event.Post.Text),
			Time:  postTime,
		},
	)
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
