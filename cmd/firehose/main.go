package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lard4/firehose-ingestor/internal/firehose"
	"github.com/lard4/firehose-ingestor/internal/kafka"
)

func main() {
	fmt.Println("Welcome to the Bluesky Firehose Ingestor.")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// graceful shutdown on ctrl+c
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("Shutting down...")
		cancel()
	}()

	client := firehose.NewClient()

	kafkaProducer := kafka.NewKafkaProducer()
	defer kafkaProducer.Close()

	runner := firehose.NewRunner(client, kafkaProducer)
	if err := runner.Start(ctx); err != nil {
		fmt.Println("Error running firehose client:", err)
	}

}
