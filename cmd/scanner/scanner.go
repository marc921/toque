package main

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
	"marcbrun.io/toque/pkg"
	"marcbrun.io/toque/pkg/messagebroker"

	"github.com/Ullaakut/nmap/v3"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go pkg.OnSignal(cancel)

	config, err := ParseConfig(ctx)
	if err != nil {
		zap.L().Fatal("ParseConfig", zap.Error(err))
	}

	logger := pkg.NewLogger(config.Env, "scanner")

	logger.Info("Starting...")
	defer logger.Info("Shutting down.")

	// Equivalent to `/usr/local/bin/nmap -p 80,443,843 google.com facebook.com youtube.com`,
	scanner, err := nmap.NewScanner(
		ctx,
		nmap.WithTargets("google.com", "facebook.com", "youtube.com"),
		nmap.WithPorts("80,443,843"),
	)
	if err != nil {
		logger.Fatal("unable to create nmap scanner", zap.Error(err))
	}

	result, warnings, err := scanner.Run()
	if len(*warnings) > 0 {
		logger.Warn("run finished with warnings", zap.Strings("warnings", *warnings)) // Warnings are non-critical errors from nmap.
	}
	if err != nil {
		logger.Fatal("unable to run nmap scan", zap.Error(err))
	}

	// // Create RabbitMQ publisher
	publisher, err := messagebroker.NewRabbitMQClient(
		ctx,
		logger.With(zap.String("component", "RabbitMQPublisher")),
		config.RabbitMQ.URL,
		"worker-input",
	)
	if err != nil {
		logger.Fatal("failed to create RabbitMQ publisher", zap.Error(err))
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
			logger.Error("failed to marshal nmapHost", zap.Error(err))
			continue
		}
		err = publisher.Publish(ctx, body)
		if err != nil {
			logger.Error("failed to publish a message", zap.Error(err))
		}
	}

	logger.Info(
		"Nmap done",
		zap.Int("hosts_up", len(result.Hosts)),
		zap.Float32("elapsed_seconds", result.Stats.Finished.Elapsed),
	)
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
