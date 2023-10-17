package utils

import probing "github.com/prometheus-community/pro-bing"

//go:generate mockery --name=HostPinger

type HostPinger interface {
	Ping(name string, arg ...string) bool
}

type Pinger struct{}

func (p Pinger) Ping(name string, arg ...string) bool {
	pinger, err := probing.NewPinger(name)
	if err != nil {
		return false
	}
	pinger.Count = 3
	err = pinger.Run()
	if err != nil {
		return false
	}

	return true
}
