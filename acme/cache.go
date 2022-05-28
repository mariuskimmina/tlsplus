package acme

import (
	"sync"
	"time"
)


// In memory cache
type Cache struct {
    options CacheOptions
    cache map[string]Certificate
    cacheIndex map[string][]string
    mu sync.RWMutex
    stopChan chan struct{}
    doneChan chan struct{}
}

type CacheOptions struct {
    RenewalCheckInterval time.Duration
    Capacity int
}



