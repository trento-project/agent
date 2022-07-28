package hosts

type DiscoveredHost struct {
	SSHAddress         string   `json:"ssh_address"`
	OSVersion          string   `json:"os_version"`
	HostIPAddresses    []string `json:"ip_addresses"`
	HostName           string   `json:"hostname"`
	CPUCount           int      `json:"cpu_count"`
	SocketCount        int      `json:"socket_count"`
	TotalMemoryMB      int      `json:"total_memory_mb"`
	AgentVersion       string   `json:"agent_version"`
	InstallationSource string   `json:"installation_source"`
}
