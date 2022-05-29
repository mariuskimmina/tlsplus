package acme

type Config struct {
	RenewalWindowRatio float64

	ServerName string

	Storage Storage
}
