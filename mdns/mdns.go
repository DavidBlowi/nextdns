package mdns

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/grandcat/zeroconf"
)

type Resolver struct {
	mu sync.RWMutex
	m  map[string]string
}

func (rs *Resolver) Start(ctx context.Context) error {
	r, err := zeroconf.NewResolver(nil)
	if err != nil {
		return fmt.Errorf("mdns resolver: %v", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			rs.mu.Lock()
			if rs.m == nil {
				rs.m = map[string]string{}
			}
			name := entry.HostName
			if idx := strings.IndexByte(name, '.'); idx != -1 {
				name = name[:idx] // remove .local. suffix
			}
			for _, ip := range entry.AddrIPv4 {
				rs.m[ip.String()] = name
			}
			for _, ip := range entry.AddrIPv6 {
				rs.m[ip.String()] = name
			}
			rs.mu.Unlock()
		}
	}(entries)

	if err = r.Browse(ctx, "_tcp", "local.", entries); err != nil {
		return fmt.Errorf("mdns browse: %v", err)
	}

	return nil
}

func (rs *Resolver) Lookup(ip net.IP) string {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.m[ip.String()]
}
