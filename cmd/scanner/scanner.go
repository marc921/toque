package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"marcbrun.io/toque/pkg"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go pkg.OnSignal(cancel)

	// Create RabbitMQ publisher
	publisher, err := pkg.NewRabbitMQClient()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create RabbitMQ publisher: %w", err))
	}
	defer publisher.Close()

	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ctx.Done():
			if ctx.Err() != context.Canceled {
				log.Printf("context error: %v", ctx.Err())
			}
			return
		case <-ticker.C:
			err = publisher.Publish(ctx, fmt.Sprint(time.Now().UnixNano()))
			if err != nil {
				log.Printf("failed to publish a message: %v", err)
			}
		}
	}

}
