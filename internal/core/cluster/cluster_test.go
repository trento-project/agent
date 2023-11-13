//nolint:exhaustruct
package cluster_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cluster"
	"github.com/trento-project/agent/internal/core/cluster/cib"
	"github.com/trento-project/agent/internal/core/cluster/crmmon"
	"github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type ClusterTestSuite struct {
	suite.Suite
}

func TestClusterTestSuite(t *testing.T) {
	suite.Run(t, new(ClusterTestSuite))
}

func (suite *ClusterTestSuite) TestNewClusterWithDiscoveryTools() {
	mockCommand := new(mocks.CommandExecutor)
	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").
		Return([]byte("7783-7084-3265-9085-8269-3286-77"), nil)
	mockCommand.On("Exec", "/usr/sbin/sbd", "-d", "/dev/vdb", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Exec", "/usr/sbin/sbd", "-d", "/dev/vdb", "list").Return(mockSbdList(), nil)
	mockCommand.On("Exec", "/usr/sbin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Exec", "/usr/sbin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdList(), nil)

	c, err := cluster.NewClusterWithDiscoveryTools(&cluster.DiscoveryTools{
		CibAdmPath:      helpers.GetFixturePath("discovery/cluster/fake_cibadmin.sh"),
		CrmmonAdmPath:   helpers.GetFixturePath("discovery/cluster/fake_crm_mon.sh"),
		CorosyncKeyPath: helpers.GetFixturePath("discovery/cluster/authkey"),
		SBDPath:         "/usr/sbin/sbd",
		SBDConfigPath:   helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"),
		CommandExecutor: mockCommand,
	})

	suite.Equal("hana_cluster", c.Name)
	suite.Equal("47d1190ffb4f781974c8356d7f863b03", c.ID)
	suite.Equal(false, c.DC)
	suite.Equal("azure", c.Provider)
	suite.Equal("/dev/vdc;/dev/vdb", c.SBD.Config["SBD_DEVICE"])
	suite.NoError(err)
}

func (suite *ClusterTestSuite) TestClusterName() {
	root := new(cib.Root)

	crmConfig := struct {
		ClusterProperties []cib.Attribute `xml:"cluster_property_set>nvpair"`
	}{
		ClusterProperties: []cib.Attribute{
			{
				ID:    "cib-bootstrap-options-cluster-name",
				Value: "cluster_name",
			},
		},
	}

	root.Configuration.CrmConfig = crmConfig

	c := cluster.Cluster{
		Cib: *root,
		Crmmon: crmmon.Root{
			Version: "1.2.3",
			Nodes: []crmmon.Node{
				{
					Name: "othernode",
					DC:   false,
				},
				{
					Name: "yetanothernode",
					DC:   true,
				},
			},
		},
		Name: "cluster_name",
	}

	suite.Equal("cluster_name", c.Name)
}

func (suite *ClusterTestSuite) TestIsDC() {
	host, _ := os.Hostname()
	root := new(cib.Root)

	crmConfig := struct {
		ClusterProperties []cib.Attribute `xml:"cluster_property_set>nvpair"`
	}{
		ClusterProperties: []cib.Attribute{
			{
				ID:    "cib-bootstrap-options-cluster-name",
				Value: "cluster_name",
			},
		},
	}

	root.Configuration.CrmConfig = crmConfig

	c := &cluster.Cluster{
		Cib: *root,
		Crmmon: crmmon.Root{
			Version: "1.2.3",
			Nodes: []crmmon.Node{
				{
					Name: "othernode",
					DC:   false,
				},
				{
					Name: host,
					DC:   true,
				},
			},
		},
	}

	suite.Equal(true, c.IsDC())

	c = &cluster.Cluster{
		Cib: *root,
		Crmmon: crmmon.Root{
			Version: "1.2.3",
			Nodes: []crmmon.Node{
				{
					Name: "othernode",
					DC:   true,
				},
				{
					Name: host,
					DC:   false,
				},
			},
		},
	}

	suite.Equal(false, c.IsDC())
}

func (suite *ClusterTestSuite) TestIsFencingEnabled() {
	root := new(cib.Root)

	crmConfig := struct {
		ClusterProperties []cib.Attribute `xml:"cluster_property_set>nvpair"`
	}{
		ClusterProperties: []cib.Attribute{
			{
				ID:    "cib-bootstrap-options-stonith-enabled",
				Value: "true",
			},
		},
	}

	root.Configuration.CrmConfig = crmConfig

	c := cluster.Cluster{
		Cib: *root,
	}

	suite.Equal(true, c.IsFencingEnabled())

	crmConfig = struct {
		ClusterProperties []cib.Attribute `xml:"cluster_property_set>nvpair"`
	}{
		ClusterProperties: []cib.Attribute{
			{
				ID:    "cib-bootstrap-options-stonith-enabled",
				Value: "false",
			},
		},
	}

	root.Configuration.CrmConfig = crmConfig

	c = cluster.Cluster{
		Cib: *root,
	}

	suite.Equal(false, c.IsFencingEnabled())
}

func (suite *ClusterTestSuite) TestFencingType() {
	c := cluster.Cluster{
		Crmmon: crmmon.Root{
			Version: "1.2.3",
			Resources: []crmmon.Resource{
				{
					Agent: "stonith:myfencing",
				},
			},
		},
	}

	suite.Equal("myfencing", c.FencingType())

	c = cluster.Cluster{
		Crmmon: crmmon.Root{
			Version: "1.2.3",
			Resources: []crmmon.Resource{
				{
					Agent: "notstonith:myfencing",
				},
			},
		},
	}

	suite.Equal("notconfigured", c.FencingType())
}

func (suite *ClusterTestSuite) TestFencingResourceExists() {
	c := cluster.Cluster{
		Crmmon: crmmon.Root{
			Version: "1.2.3",
			Resources: []crmmon.Resource{
				{
					Agent: "stonith:myfencing",
				},
			},
		},
	}

	suite.Equal(true, c.FencingResourceExists())

	c = cluster.Cluster{
		Crmmon: crmmon.Root{
			Version: "1.2.3",
			Resources: []crmmon.Resource{
				{
					Agent: "notstonith:myfencing",
				},
			},
		},
	}

	suite.Equal(false, c.FencingResourceExists())
}

func (suite *ClusterTestSuite) TestIsFencingSBD() {
	c := cluster.Cluster{
		Crmmon: crmmon.Root{
			Version: "1.2.3",
			Resources: []crmmon.Resource{
				{
					Agent: "stonith:external/sbd",
				},
			},
		},
	}

	suite.Equal(true, c.IsFencingSBD())

	c = cluster.Cluster{
		Crmmon: crmmon.Root{
			Version: "1.2.3",
			Resources: []crmmon.Resource{
				{
					Agent: "stonith:other",
				},
			},
		},
	}

	suite.Equal(false, c.IsFencingSBD())
}
