package mocks

import (
	"github.com/trento-project/agent/internal/cloud"
	"github.com/trento-project/agent/internal/cluster"
	"github.com/trento-project/agent/test/helpers"
)

func NewDiscoveredClusterMock() cluster.Cluster {
	cluster, _ := cluster.NewClusterWithDiscoveryTools(&cluster.DiscoveryTools{
		CibAdmPath:      helpers.GetFixtureFile("discovery/cluster/fake_cibadmin.sh"),
		CrmmonAdmPath:   helpers.GetFixtureFile("discovery/cluster/fake_crm_mon.sh"),
		CorosyncKeyPath: helpers.GetFixtureFile("discovery/cluster/authkey"),
		SBDPath:         helpers.GetFixtureFile("discovery/cluster/fake_sbd.sh"),
		SBDConfigPath:   helpers.GetFixtureFile("discovery/cluster/sbd/sbd_config"),
	})

	cluster.Provider = cloud.Azure

	return cluster
}
