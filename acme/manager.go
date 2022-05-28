package acme

import (
	"fmt"

	"github.com/mholt/acmez"
)

type AcmeManager struct {
    CA string
    Email string
    DNS01Solver acmez.Solver
    Config *Config
}


func NewACMEManager(cfg *Config) (*AcmeManager, error) {
    if cfg == nil {
        return nil, fmt.Errorf("Missing Config")
    }

    return &AcmeManager{
        CA: "Let's Encrypt",
        Email: "Test@test.test",
        DNS01Solver: &DNSSolver{},
        Config: cfg,
    }, nil
}
