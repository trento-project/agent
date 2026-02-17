package mocks

import (
	"github.com/trento-project/agent/internal/core/hosts"
	"github.com/trento-project/agent/internal/core/systemd"
)

func NewDiscoveredHostMock() hosts.DiscoveredHost {
	fqdn := "com.example.trento.host"

	return hosts.DiscoveredHost{
		OSVersion:                "15-SP2",
		Architecture:             "x86_64",
		HostIPAddresses:          []string{"10.1.1.4", "10.1.1.5", "10.1.1.6"},
		Netmasks:                 []int{24, 16, 32},
		HostName:                 "thehostnamewherethediscoveryhappened",
		CPUCount:                 2,
		SocketCount:              1,
		TotalMemoryMB:            4096,
		AgentVersion:             "trento-agent-version",
		InstallationSource:       "Community",
		FullyQualifiedDomainName: &fqdn,
		PrometheusTargets: map[string]string{
			"node_exporter": "10.1.1.4:9100",
		},
		PrometheusMode: "pull",
		SystemdUnits: []systemd.UnitInfo{
			{
				Name:          "pacemaker.service",
				UnitFileState: "enabled",
			},
			{
				Name:          "another.service",
				UnitFileState: "disabled",
			},
			{
				Name:          "yet.another.service",
				UnitFileState: "unknown",
			},
		},
	}
}
