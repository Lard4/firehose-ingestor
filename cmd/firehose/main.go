package main

import (
	"context"
	"fmt"
	"github.com/lard4/firehose-ingestor/internal/firehose"
	"os"
	"os/signal"
	"syscall"
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
	runner := firehose.NewRunner(client)
	if err := runner.Run(ctx); err != nil {
		fmt.Println("Error running firehose client:", err)
	}

}
