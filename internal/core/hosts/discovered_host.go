package hosts

type NetworkInterface struct {
	Index     int       `json:"index"`
	Name      string    `json:"name"`
	Addresses []Address `json:"addresses"`
}

type Address struct {
	Address string `json:"address"`
	Netmask int    `json:"netmask"`
}

type DiscoveredHost struct {
	OSVersion                string             `json:"os_version"`
	HostIPAddresses          []string           `json:"ip_addresses"` // deprecated
	NetworkInterfaces        []NetworkInterface `json:"network_interfaces"`
	HostName                 string             `json:"hostname"`
	CPUCount                 int                `json:"cpu_count"`
	SocketCount              int                `json:"socket_count"`
	TotalMemoryMB            int                `json:"total_memory_mb"`
	AgentVersion             string             `json:"agent_version"`
	InstallationSource       string             `json:"installation_source"`
	FullyQualifiedDomainName *string            `json:"fully_qualified_domain_name,omitempty"`
}
