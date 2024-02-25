package main

import (
	"fmt"
	"log"

	"marcbrun.io/kubernetes-api/pkg"
)

func main() {
	// Create RabbitMQ publisher
	consumer, err := pkg.NewRabbitMQClient()
	if err != nil {
		log.Fatal(fmt.Errorf("pkg.NewRabbitMQClient: %w", err))
	}
	defer consumer.Close()

	msgs, err := consumer.Consume()
	if err != nil {
		log.Fatal(fmt.Errorf("consumer.Consume: %w", err))
	}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			err = d.Ack(false)
			if err != nil {
				log.Printf("Failed to acknowledge the message: %v", err)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	pkg.OnSignal(func() {
		err = consumer.Close()
		if err != nil {
			log.Printf("Failed to close the consumer: %v", err)
		}
	})
}
