package hosts_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/hosts"
)

func TestDiscoveredHost(t *testing.T) {
	suite.Run(t, new(DiscoveredHostTestSuite))
}

type DiscoveredHostTestSuite struct {
	suite.Suite
}

func (suite *DiscoveredHostTestSuite) TestUTCTimeMarshalJSON() {
	utcTime := hosts.UTCTime{Time: time.Date(2024, 6, 10, 15, 30, 0, 0, time.UTC)}
	data, err := utcTime.MarshalJSON()
	suite.NoError(err)
	suite.NotEmpty(data)
	suite.Equal(`"2024-06-10T15:30:00Z"`, string(data))
}

func (suite *DiscoveredHostTestSuite) TestDiscoveredHostMarshalJSON() {
	fqdn := "test-host.local"
	host := hosts.DiscoveredHost{
		OSVersion:                "openSUSE Leap 15.3",
		Architecture:             "x86_64",
		HostIPAddresses:          []string{"192.168.1.10"},
		Netmasks:                 []int{24},
		HostName:                 "test-host",
		CPUCount:                 4,
		SocketCount:              1,
		TotalMemoryMB:            8192,
		AgentVersion:             "1.0.0",
		InstallationSource:       "package",
		FullyQualifiedDomainName: &fqdn,
		PrometheusMode:           "pull",
		LastBootTimestamp:        &hosts.UTCTime{Time: time.Date(2024, 6, 10, 15, 30, 0, 0, time.UTC)},
	}

	data, err := json.Marshal(host)
	suite.NoError(err)
	suite.NotEmpty(data)

	expected := strings.ReplaceAll(`{"os_version":"openSUSE Leap 15.3","arch":"x86_64","ip_addresses":["192.168.1.10"],"netmasks":[24],
"hostname":"test-host","cpu_count":4,"socket_count":1,"total_memory_mb":8192,"agent_version":"1.0.0",
"installation_source":"package","fully_qualified_domain_name":"test-host.local","prometheus_targets":null,
"prometheus_mode":"pull","systemd_units":null,"last_boot_timestamp":"2024-06-10T15:30:00Z"}`, "\n", "")

	suite.Equal(expected, string(data))
}
