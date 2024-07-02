package net

import (
	"fmt"
	stdnet "net"
)

// Lookup resolves hostnames to IP addresses.
func Lookup(host string) ([]string, error) {
	if host == "" {
		return nil, fmt.Errorf("lookup host: empty host")
	}
	ips, err := stdnet.LookupHost(host)
	if err != nil {
		return nil, fmt.Errorf("lookup host: %w", err)
	}
	return ips, nil
}
