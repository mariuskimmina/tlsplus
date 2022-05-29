package acme

import (
	"context"
	"fmt"

	"github.com/mholt/acmez"
)

type AcmeManager struct {
	CA          string
	Email       string
	DNS01Solver acmez.Solver
	Config      *Config
}

func NewACMEManager(cfg *Config) (*AcmeManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Missing Config")
	}

	return &AcmeManager{
		CA:          "https://127.0.0.1:14000/dir", //pebble
		Email:       "Test@test.test",
		DNS01Solver: &DNSSolver{},
		Config:      cfg,
	}, nil
}


func (AcmeManager) renewManagedCertificates(ctx context.Context) error {

}
