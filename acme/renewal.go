package acme

import (
	"context"
	"fmt"
	"time"
)

// an indefinitely looping function that, on a regular schedule, checks certificates for expiration
// and initiates the renewal of certs that are expiring soon
func (m *AcmeManager) RenewalLoop() {
    fmt.Println("Starting to manage certificates in the background")
    renewalTicker := time.NewTicker(m.Config.RenewCheckInterval)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    for {
		select {
		case <-renewalTicker.C:
			err := m.renewManagedCertificates(ctx)
			if err != nil {
                fmt.Println("Error during certificate renewal: %v", err)
			}
        }
    }
    fmt.Println("RenewalLoop has ended")
}
