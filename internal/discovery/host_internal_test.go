package discovery

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type HostInternalTestSuite struct {
	suite.Suite
	ipAddresses []string
}

func TestHostInternalTestSuite(t *testing.T) {
	suite.Run(t, new(HostInternalTestSuite))
}

func (suite *HostInternalTestSuite) SetupSuite() {
	suite.ipAddresses = []string{"127.0.0.1", "::1", "10.1.1.5", "10.1.1.4", "10.1.1.6", "6c62:7cc9:3936:e802:2bbe"}
}

func (suite *HostInternalTestSuite) TestUpdatePrometheusTargets() {
	expectedTargets := PrometheusTargets{
		"node_exporter": "10.1.1.4:9100",
	}

	updatedTargets := updatePrometheusTargets("", suite.ipAddresses, "node_exporter")
	suite.Equal(expectedTargets, updatedTargets)
}

func (suite *HostInternalTestSuite) TestUpdatePrometheusTargetsGivenByUser() {
	expectedTargets := PrometheusTargets{
		"node_exporter": "192.168.1.60:9123",
	}

	updatedTargets := updatePrometheusTargets("192.168.1.60:9123", suite.ipAddresses, "node_exporter")
	suite.Equal(expectedTargets, updatedTargets)
}

func (suite *HostInternalTestSuite) TestUpdatePrometheusTargetsCustomName() {
	expectedTargets := PrometheusTargets{
		"custom_exporter": "192.168.1.60:9123",
	}

	updatedTargets := updatePrometheusTargets("192.168.1.60:9123", suite.ipAddresses, "custom_exporter")
	suite.Equal(expectedTargets, updatedTargets)
}

func (suite *HostInternalTestSuite) TestHostLastBootTime() {
	lastBootTimestamp := getLastBootTimestamp()
	suite.NotNil(lastBootTimestamp)
	suite.Less(lastBootTimestamp.Unix(), int64(9999999999))
}
