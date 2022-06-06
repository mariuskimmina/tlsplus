package acme

import "time"

const (
	// DefaultRenewCheckInterval is how often to check certificates for expiration.
	// Scans are very lightweight, so this can be semi-frequent. This default should
	// be smaller than <Minimum Cert Lifetime>*DefaultRenewalWindowRatio/3, which
	// gives certificates plenty of chance to be renewed on time.
	DefaultRenewCheckInterval = 10 * time.Minute

	// DefaultRenewalWindowRatio is how much of a certificate's lifetime becomes the
	// renewal window. The renewal window is the span of time at the end of the
	// certificate's validity period in which it should be renewed. A default value
	// of ~1/3 is pretty safe and recommended for most certificates.
	DefaultRenewalWindowRatio = 1.0 / 3.0
)
