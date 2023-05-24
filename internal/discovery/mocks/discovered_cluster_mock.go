package mocks

import (
	"github.com/trento-project/agent/internal/core/cloud"
	"github.com/trento-project/agent/internal/core/cluster"
	"github.com/trento-project/agent/test/helpers"
)

func NewDiscoveredClusterMock() *cluster.Cluster {
	cluster, _ := cluster.NewClusterWithDiscoveryTools(&cluster.DiscoveryTools{
		CibAdmPath:      helpers.GetFixturePath("discovery/cluster/fake_cibadmin.sh"),
		CrmmonAdmPath:   helpers.GetFixturePath("discovery/cluster/fake_crm_mon.sh"),
		CorosyncKeyPath: helpers.GetFixturePath("discovery/cluster/authkey"),
		SBDPath:         helpers.GetFixturePath("discovery/cluster/fake_sbd.sh"),
		SBDConfigPath:   helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"),
	})

	cluster.Provider = cloud.Azure

	return cluster
}
