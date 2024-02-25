package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgtype"
	"marcbrun.io/kubernetes-api/db/sqlcgen"
	"marcbrun.io/kubernetes-api/pkg"
)

func main() {
	// Connect to the database.
	ctx := context.Background()
	connString, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	dbConn, err := pkg.NewPostgresConnection(ctx, connString)
	if err != nil {
		log.Fatal(fmt.Errorf("pkg.NewPostgresConnection: %w", err))
	}
	defer dbConn.Close(ctx)

	queries := sqlcgen.New(dbConn)

	// Create RabbitMQ publisher
	msgConsumer, err := pkg.NewRabbitMQClient()
	if err != nil {
		log.Fatal(fmt.Errorf("pkg.NewRabbitMQClient: %w", err))
	}
	defer msgConsumer.Close()

	msgs, err := msgConsumer.Consume()
	if err != nil {
		log.Fatal(fmt.Errorf("consumer.Consume: %w", err))
	}

	go func() {
		for d := range msgs {
			if len(d.Body) == 0 {
				log.Printf("Empty message, skipping")
				continue
			}
			log.Printf("Received a message: %s", d.Body)
			msg, err := queries.InsertMessage(ctx, pgtype.Text{
				String: string(d.Body),
				Valid:  true,
			})
			if err != nil {
				log.Printf("Failed to insert message: %v", err)
			} else {
				log.Printf("Inserted message: %#v", msg)
			}
			err = d.Ack(false)
			if err != nil {
				log.Printf("Failed to acknowledge the message: %v", err)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	pkg.OnSignal(func() {
		err = msgConsumer.Close()
		if err != nil {
			log.Printf("Failed to close the consumer: %v", err)
		}
	})
}
