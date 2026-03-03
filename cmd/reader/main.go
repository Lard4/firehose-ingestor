package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lard4/firehose-ingestor/internal/consumers"
)

func main() {
	fmt.Println("Welcome to the Bluesky Post Reader.")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaConsumer := consumers.NewKafkaConsumer()
	defer kafkaConsumer.Close()

	// graceful shutdown on ctrl+c
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("Shutting down...")
		kafkaConsumer.Close()
		cancel()
	}()

	kafkaConsumer.Start(ctx)
}
