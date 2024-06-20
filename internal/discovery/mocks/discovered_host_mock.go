package mocks

import "github.com/trento-project/agent/internal/core/hosts"

func NewDiscoveredHostMock() hosts.DiscoveredHost {
	fqdn := "com.example.trento.host"

	return hosts.DiscoveredHost{
		OSVersion:                "15-SP2",
		HostIPAddresses:          []string{"10.1.1.4", "10.1.1.5", "10.1.1.6"},
		HostIPAddressesNetmasks:  []string{"10.1.1.4/16", "10.1.1.5/24", "10.1.1.6/32"},
		HostName:                 "thehostnamewherethediscoveryhappened",
		CPUCount:                 2,
		SocketCount:              1,
		TotalMemoryMB:            4096,
		AgentVersion:             "trento-agent-version",
		InstallationSource:       "Community",
		FullyQualifiedDomainName: &fqdn,
	}
}
