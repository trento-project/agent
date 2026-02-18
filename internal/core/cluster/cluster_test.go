//nolint:exhaustruct
package cluster_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cluster"
	"github.com/trento-project/agent/internal/core/cluster/cib"
	"github.com/trento-project/agent/internal/core/cluster/crmmon"
	mocksCluster "github.com/trento-project/agent/internal/core/cluster/mocks"
	mocksSystemd "github.com/trento-project/agent/internal/core/systemd/mocks"
	"github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

const (
	NotExistingFile = "not-existing"
)

type ClusterTestSuite struct {
	suite.Suite
}

func TestClusterTestSuite(t *testing.T) {
	suite.Run(t, new(ClusterTestSuite))
}

func (suite *ClusterTestSuite) TestNewClusterWithDiscoveryTools() {
	ctx := context.Background()
	mockCommand := new(mocks.MockCommandExecutor)
	mockCommand.On("Output", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return([]byte("7783-7084-3265-9085-8269-3286-77"), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdb", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdb", "list").Return(mockSbdList(), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdList(), nil)

	mockSystemd := new(mocksSystemd.MockSystemd)
	mockSystemd.On("IsActive", ctx, "corosync.service").Return(true, nil)
	mockSystemd.On("IsActive", ctx, "pacemaker.service").Return(true, nil)

	mockCmdClient := new(mocksCluster.MockCmdClient)
	mockCmdClient.On("GetState", ctx).Return("S_IDLE", nil)

	c, err := cluster.NewClusterWithDiscoveryTools(ctx, &cluster.DiscoveryTools{
		CibAdmPath:         helpers.GetFixturePath("discovery/cluster/fake_cibadmin.sh"),
		CrmmonAdmPath:      helpers.GetFixturePath("discovery/cluster/fake_crm_mon.sh"),
		CorosyncKeyPath:    helpers.GetFixturePath("discovery/cluster/authkey"),
		CorosyncConfigPath: helpers.GetFixturePath("discovery/cluster/corosync.conf"),
		SBDPath:            "/usr/sbin/sbd",
		SBDConfigPath:      helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"),
		CommandExecutor:    mockCommand,
		SystemdConnector:   mockSystemd,
		CmdClient:          mockCmdClient,
	})

	suite.Equal("hana_cluster", c.Name)
	suite.Equal("47d1190ffb4f781974c8356d7f863b03", c.ID)
	suite.Equal(false, c.DC)
	suite.Equal("azure", c.Provider)
	suite.Equal("/dev/vdc;/dev/vdb", c.SBD.Config["SBD_DEVICE"])
	suite.Equal("idle", c.State)
	suite.NoError(err)
}

func (suite *ClusterTestSuite) TestNewClusterDisklessSBD() {
	ctx := context.Background()
	mockCommand := new(mocks.MockCommandExecutor)
	mockCommand.On("Output", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return([]byte("7783-7084-3265-9085-8269-3286-77"), nil)

	mockSystemd := new(mocksSystemd.MockSystemd)
	mockSystemd.On("IsActive", ctx, "corosync.service").Return(true, nil)
	mockSystemd.On("IsActive", ctx, "pacemaker.service").Return(true, nil)

	mockCmdClient := new(mocksCluster.MockCmdClient)
	mockCmdClient.On("GetState", ctx).Return("S_IDLE", nil)

	c, err := cluster.NewClusterWithDiscoveryTools(ctx, &cluster.DiscoveryTools{
		CibAdmPath:         helpers.GetFixturePath("discovery/cluster/fake_cibadmin.sh"),
		CrmmonAdmPath:      helpers.GetFixturePath("discovery/cluster/fake_crm_mon_diskless_sbd.sh"),
		CorosyncKeyPath:    helpers.GetFixturePath("discovery/cluster/authkey"),
		CorosyncConfigPath: helpers.GetFixturePath("discovery/cluster/corosync.conf"),
		SBDPath:            "/usr/sbin/sbd",
		SBDConfigPath:      helpers.GetFixturePath("discovery/cluster/sbd/sbd_config_no_device"),
		CommandExecutor:    mockCommand,
		SystemdConnector:   mockSystemd,
		CmdClient:          mockCmdClient,
	})

	suite.Equal("hana_cluster", c.Name)
	suite.Equal("47d1190ffb4f781974c8356d7f863b03", c.ID)
	suite.Equal(false, c.DC)
	suite.Equal("azure", c.Provider)
	suite.Equal("/dev/watchdog", c.SBD.Config["SBD_WATCHDOG_DEV"])
	suite.Equal([]*cluster.SBDDevice(nil), c.SBD.Devices)
	suite.Equal(true, c.Online)
	suite.NoError(err)
}

//nolint:dupl
func (suite *ClusterTestSuite) TestNewClusterWithOfflineHost() {
	ctx := context.Background()
	mockCommand := new(mocks.MockCommandExecutor)
	mockCommand.On("Output", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return([]byte("7783-7084-3265-9085-8269-3286-77"), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdb", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdb", "list").Return(mockSbdList(), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdList(), nil)

	mockSystemd := new(mocksSystemd.MockSystemd)
	mockSystemd.On("IsActive", ctx, "corosync.service").Return(true, nil)
	mockSystemd.On("IsActive", ctx, "pacemaker.service").Return(false, nil)

	c, err := cluster.NewClusterWithDiscoveryTools(ctx, &cluster.DiscoveryTools{
		CibAdmPath:         helpers.GetFixturePath("discovery/cluster/fake_cibadmin.sh"),
		CrmmonAdmPath:      helpers.GetFixturePath("discovery/cluster/fake_crm_mon.sh"),
		CorosyncKeyPath:    helpers.GetFixturePath("discovery/cluster/authkey"),
		CorosyncConfigPath: helpers.GetFixturePath("discovery/cluster/corosync.conf"),
		SBDPath:            "/usr/sbin/sbd",
		SBDConfigPath:      helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"),
		CommandExecutor:    mockCommand,
		SystemdConnector:   mockSystemd,
		CmdClient:          new(mocksCluster.MockCmdClient),
	})

	suite.Equal("hana_cluster", c.Name)
	suite.Equal("47d1190ffb4f781974c8356d7f863b03", c.ID)
	suite.Equal(false, c.Online)
	suite.NoError(err)
}

//nolint:dupl
func (suite *ClusterTestSuite) TestNewClusterWithOfflineHostNoName() {
	ctx := context.Background()
	mockCommand := new(mocks.MockCommandExecutor)
	mockCommand.On("Output", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return([]byte("7783-7084-3265-9085-8269-3286-77"), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdb", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdb", "list").Return(mockSbdList(), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdList(), nil)

	mockSystemd := new(mocksSystemd.MockSystemd)
	mockSystemd.On("IsActive", ctx, "corosync.service").Return(false, errors.New(""))
	mockSystemd.On("IsActive", ctx, "pacemaker.service").Return(false, errors.New(""))

	c, err := cluster.NewClusterWithDiscoveryTools(ctx, &cluster.DiscoveryTools{
		CibAdmPath:         helpers.GetFixturePath("discovery/cluster/fake_cibadmin.sh"),
		CrmmonAdmPath:      helpers.GetFixturePath("discovery/cluster/fake_crm_mon.sh"),
		CorosyncKeyPath:    helpers.GetFixturePath("discovery/cluster/authkey"),
		CorosyncConfigPath: helpers.GetFixturePath("discovery/cluster/corosync_noname.conf"),
		SBDPath:            "/usr/sbin/sbd",
		SBDConfigPath:      helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"),
		CommandExecutor:    mockCommand,
		SystemdConnector:   mockSystemd,
		CmdClient:          new(mocksCluster.MockCmdClient),
	})

	suite.Equal("", c.Name)
	suite.Equal("47d1190ffb4f781974c8356d7f863b03", c.ID)
	suite.Equal(false, c.Online)
	suite.NoError(err)
}

func (suite *ClusterTestSuite) TestNewClusterWithUnknownState() {
	ctx := context.Background()
	mockCommand := new(mocks.MockCommandExecutor)
	mockCommand.On("Output", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return([]byte("7783-7084-3265-9085-8269-3286-77"), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdb", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdb", "list").Return(mockSbdList(), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Output", "/usr/sbin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdList(), nil)

	mockSystemd := new(mocksSystemd.MockSystemd)
	mockSystemd.On("IsActive", ctx, "corosync.service").Return(true, nil)
	mockSystemd.On("IsActive", ctx, "pacemaker.service").Return(true, nil)

	mockCmdClient := new(mocksCluster.MockCmdClient)
	mockCmdClient.On("GetState", ctx).Return("", errors.New(""))

	c, err := cluster.NewClusterWithDiscoveryTools(ctx, &cluster.DiscoveryTools{
		CibAdmPath:         helpers.GetFixturePath("discovery/cluster/fake_cibadmin.sh"),
		CrmmonAdmPath:      helpers.GetFixturePath("discovery/cluster/fake_crm_mon.sh"),
		CorosyncKeyPath:    helpers.GetFixturePath("discovery/cluster/authkey"),
		CorosyncConfigPath: helpers.GetFixturePath("discovery/cluster/corosync.conf"),
		SBDPath:            "/usr/sbin/sbd",
		SBDConfigPath:      helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"),
		CommandExecutor:    mockCommand,
		SystemdConnector:   mockSystemd,
		CmdClient:          mockCmdClient,
	})

	suite.Equal("unknown", c.State)
	suite.NoError(err)
}

func (suite *ClusterTestSuite) TestNewClusterCorosyncNotConfigured() {
	ctx := context.Background()
	c, err := cluster.NewClusterWithDiscoveryTools(ctx, &cluster.DiscoveryTools{
		CorosyncConfigPath: NotExistingFile,
		CibAdmPath:         helpers.GetFixturePath("discovery/cluster/fake_cibadmin.sh"),
		CrmmonAdmPath:      helpers.GetFixturePath("discovery/cluster/fake_crm_mon_diskless_sbd.sh"),
		CorosyncKeyPath:    helpers.GetFixturePath("discovery/cluster/authkey"),
		SBDPath:            "/usr/sbin/sbd",
		SBDConfigPath:      helpers.GetFixturePath("discovery/cluster/sbd/sbd_config_no_device"),
		CommandExecutor:    new(mocks.MockCommandExecutor),
		SystemdConnector:   new(mocksSystemd.MockSystemd),
		CmdClient:          new(mocksCluster.MockCmdClient),
	})

	suite.Nil(c)
	suite.Error(err)
}

func (suite *ClusterTestSuite) TestNewClusterCorosyncNoAuthkeyConfigured() {
	ctx := context.Background()
	c, err := cluster.NewClusterWithDiscoveryTools(ctx, &cluster.DiscoveryTools{
		CorosyncKeyPath:    NotExistingFile,
		CibAdmPath:         helpers.GetFixturePath("discovery/cluster/fake_cibadmin.sh"),
		CrmmonAdmPath:      helpers.GetFixturePath("discovery/cluster/fake_crm_mon_diskless_sbd.sh"),
		CorosyncConfigPath: helpers.GetFixturePath("discovery/cluster/corosync.conf"),
		SBDPath:            "/usr/sbin/sbd",
		SBDConfigPath:      helpers.GetFixturePath("discovery/cluster/sbd/sbd_config_no_device"),
		CommandExecutor:    new(mocks.MockCommandExecutor),
		SystemdConnector:   new(mocksSystemd.MockSystemd),
		CmdClient:          new(mocksCluster.MockCmdClient),
	})

	suite.Nil(c)
	suite.Error(err)
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
