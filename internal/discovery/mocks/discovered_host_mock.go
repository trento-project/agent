package mocks

import "github.com/trento-project/agent/internal/core/hosts"

func NewDiscoveredHostMock() hosts.DiscoveredHost {
	fqdn := "com.example.trento.host"

	return hosts.DiscoveredHost{
		OSVersion:       "15-SP2",
		HostIPAddresses: []string{"10.1.1.4", "10.1.1.5", "10.1.1.6"},
		NetworkInterfaces: []hosts.NetworkInterface{
			{
				Index: 0,
				Name:  "eth0",
				Addresses: []hosts.Address{
					{
						Address: "10.1.1.4",
						Netmask: 24,
					},
					{
						Address: "10.1.1.5",
						Netmask: 16,
					},
					{
						Address: "10.1.1.6",
						Netmask: 32,
					},
				},
			},
			{
				Index: 1,
				Name:  "eth1",
				Addresses: []hosts.Address{
					{
						Address: "10.1.2.4",
						Netmask: 24,
					},
				},
			},
		},
		HostName:                 "thehostnamewherethediscoveryhappened",
		CPUCount:                 2,
		SocketCount:              1,
		TotalMemoryMB:            4096,
		AgentVersion:             "trento-agent-version",
		InstallationSource:       "Community",
		FullyQualifiedDomainName: &fqdn,
	}
}
