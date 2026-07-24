package helpers

import (
	"context"
	"net"
	"sync"

	"golang.org/x/sync/singleflight"
)

// CachedDNSResolver resolves host names using net.DefaultResolver and caches
// the results for the lifetime of the process. Concurrent lookups of the same
// host are deduplicated.
type CachedDNSResolver struct {
	mu    sync.RWMutex
	cache map[string][]string
	group singleflight.Group
}

// LookupHost returns the cached addresses for host, resolving and caching them
// on first use.
func (r *CachedDNSResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	r.mu.RLock()
	ips, ok := r.cache[host]
	r.mu.RUnlock()
	if ok {
		return ips, nil
	}

	result, err, _ := r.group.Do(host, func() (any, error) {
		ips, err := net.DefaultResolver.LookupHost(ctx, host)
		if err != nil {
			return nil, err
		}
		r.mu.Lock()
		if r.cache == nil {
			r.cache = make(map[string][]string)
		}
		r.cache[host] = ips
		r.mu.Unlock()
		return ips, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]string), nil
}
