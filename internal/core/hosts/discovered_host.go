package hosts

import "time"

import "github.com/trento-project/agent/internal/core/systemd"

type UTCTime struct{ time.Time }

func (t UTCTime) MarshalJSON() ([]byte, error) {
	formatted := t.Time.UTC().Format(time.RFC3339)
	return []byte(`"` + formatted + `"`), nil
}

func (t *UTCTime) UnmarshalJSON(data []byte) error {
	parsed, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return err
	}
	*t = UTCTime{parsed}
	return nil
}

type DiscoveredHost struct {
	OSVersion                string             `json:"os_version"`
	Architecture             string             `json:"arch"`
	HostIPAddresses          []string           `json:"ip_addresses"`
	Netmasks                 []int              `json:"netmasks"`
	HostName                 string             `json:"hostname"`
	CPUCount                 int                `json:"cpu_count"`
	SocketCount              int                `json:"socket_count"`
	TotalMemoryMB            uint64             `json:"total_memory_mb"`
	AgentVersion             string             `json:"agent_version"`
	InstallationSource       string             `json:"installation_source"`
	FullyQualifiedDomainName *string            `json:"fully_qualified_domain_name,omitempty"`
	PrometheusTargets        map[string]string  `json:"prometheus_targets"`
	PrometheusMode           string             `json:"prometheus_mode"`
	SystemdUnits             []systemd.UnitInfo `json:"systemd_units"`
	LastBootTimestamp        *UTCTime           `json:"last_boot_timestamp,omitempty"`
}
