package hosts

type DiscoveredHost struct {
	OSVersion                string   `json:"os_version"`
	HostIPAddresses          []string `json:"ip_addresses"`
	HostName                 string   `json:"hostname"`
	CPUCount                 int      `json:"cpu_count"`
	SocketCount              int      `json:"socket_count"`
	TotalMemoryMB            int      `json:"total_memory_mb"`
	AgentVersion             string   `json:"agent_version"`
	InstallationSource       string   `json:"installation_source"`
	FullyQualifiedDomainName *string  `json:"fully_qualified_domain_name, omitempty"`
}
