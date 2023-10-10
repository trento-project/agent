package utils

import "net"

//go:generate mockery --name=HostnameResolver

type HostnameResolver interface {
	LookupHost(host string) (addrs []string, err error)
}

type Resolver struct{}

func (r Resolver) LookupHost(host string) (addrs []string, err error) {
	return net.LookupHost(host)
}
