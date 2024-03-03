package main

import (
	"context"
	"encoding/json"
	"fmt"
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

	// // Create RabbitMQ publisher
	publisher, err := pkg.NewRabbitMQClient()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create RabbitMQ publisher: %w", err))
	}
	defer publisher.Close()

	// Use the results to print an example output
	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}

		addresses := make([]string, 0, len(host.Addresses))
		for _, address := range host.Addresses {
			addresses = append(addresses, address.Addr)
		}
		ports := make([]NmapPort, 0, len(host.Ports))
		for _, port := range host.Ports {
			ports = append(ports, NmapPort{
				ID:       port.ID,
				Protocol: port.Protocol,
				State:    port.State.State,
				Service:  port.Service.Name,
			})
		}
		nmapHost := NmapHost{
			Addresses: addresses,
			Ports:     ports,
		}
		body, err := json.Marshal(nmapHost)
		if err != nil {
			log.Printf("failed to marshal nmapHost: %v", err)
			continue
		}
		err = publisher.Publish(ctx, string(body))
		if err != nil {
			log.Printf("failed to publish a message: %v", err)
		}
	}

	log.Printf("Nmap done: %d hosts up scanned in %.2f seconds\n", len(result.Hosts), result.Stats.Finished.Elapsed)

}

type NmapHost struct {
	Addresses []string
	Ports     []NmapPort
}

type NmapPort struct {
	ID       uint16
	Protocol string
	State    string
	Service  string
}
