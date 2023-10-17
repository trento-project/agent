package utils

import "net"

//go:generate mockery --name=HostnameResolver

type HostnameResolver interface {
	LookupHost(host string) ([]string, error)
}

type Resolver struct{}

func (r Resolver) LookupHost(host string) ([]string, error) {
	return net.LookupHost(host)
}
