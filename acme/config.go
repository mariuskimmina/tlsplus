package acme

import "time"

type Config struct {
	RenewalWindowRatio float64

    RenewCheckInterval time.Duration

	ServerName string

	Storage Storage
}

func NewConfig(serverName string, storage Storage) *Config {
    return &Config{
        RenewCheckInterval: DefaultRenewCheckInterval,
        RenewalWindowRatio: DefaultRenewalWindowRatio,
        ServerName: serverName,
        Storage: storage,
    }
}
