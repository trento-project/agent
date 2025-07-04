package hosts

type DiscoveredHost struct {
	OSVersion                string            `json:"os_version"`
	Architecture             string            `json:"arch"`
	HostIPAddresses          []string          `json:"ip_addresses"`
	Netmasks                 []int             `json:"netmasks"`
	HostName                 string            `json:"hostname"`
	CPUCount                 int               `json:"cpu_count"`
	SocketCount              int               `json:"socket_count"`
	TotalMemoryMB            uint64            `json:"total_memory_mb"`
	AgentVersion             string            `json:"agent_version"`
	InstallationSource       string            `json:"installation_source"`
	FullyQualifiedDomainName *string           `json:"fully_qualified_domain_name,omitempty"`
	PrometheusTargets        map[string]string `json:"prometheus_targets"`
	SystemdUnits             []UnitInfo        `json:"systemd_units"`
}
