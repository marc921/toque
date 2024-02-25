package pkg

import (
	"os"
	"os/signal"
	"syscall"
)

// OnSignal waits (blocks) for a signal and then calls the provided function
func OnSignal(f func()) {
	sigChan := make(chan os.Signal, 1)
	// Notify the sigChan for specified signals
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	// Wait for a signal
	<-sigChan
	f()
}
