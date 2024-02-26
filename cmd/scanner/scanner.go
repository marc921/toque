package main

import (
	"context"
	"log"

	"marcbrun.io/toque/pkg"

	"github.com/Ullaakut/nmap/v3"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go pkg.OnSignal(cancel)

	// Equivalent to `/usr/local/bin/nmap -p 80,443,843 google.com facebook.com youtube.com`,
	scanner, err := nmap.NewScanner(
		ctx,
		nmap.WithTargets("google.com", "facebook.com", "youtube.com"),
		nmap.WithPorts("80,443,843"),
	)
	if err != nil {
		log.Fatalf("unable to create nmap scanner: %v", err)
	}

	result, warnings, err := scanner.Run()
	if len(*warnings) > 0 {
		log.Printf("run finished with warnings: %s\n", *warnings) // Warnings are non-critical errors from nmap.
	}
	if err != nil {
		log.Fatalf("unable to run nmap scan: %v", err)
	}

	// Use the results to print an example output
	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}

		log.Printf("Host %q:\n", host.Addresses[0])

		for _, port := range host.Ports {
			log.Printf("\tPort %d/%s %s %s\n", port.ID, port.Protocol, port.State, port.Service.Name)
		}
	}

	log.Printf("Nmap done: %d hosts up scanned in %.2f seconds\n", len(result.Hosts), result.Stats.Finished.Elapsed)

	// // Create RabbitMQ publisher
	// publisher, err := pkg.NewRabbitMQClient()
	// if err != nil {
	// 	log.Fatal(fmt.Errorf("failed to create RabbitMQ publisher: %w", err))
	// }
	// defer publisher.Close()

	// ticker := time.NewTicker(2 * time.Second)
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		if ctx.Err() != context.Canceled {
	// 			log.Printf("context error: %v", ctx.Err())
	// 		}
	// 		return
	// 	case <-ticker.C:
	// 		err = publisher.Publish(ctx, fmt.Sprint(time.Now().UnixNano()))
	// 		if err != nil {
	// 			log.Printf("failed to publish a message: %v", err)
	// 		}
	// 	}
	// }

}
