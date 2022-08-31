package mocks

import (
	"github.com/trento-project/agent/internal/cloud"
	"github.com/trento-project/agent/internal/cluster"
)

func NewDiscoveredClusterMock() cluster.Cluster {
	cluster, _ := cluster.NewClusterWithDiscoveryTools(&cluster.DiscoveryTools{
		CibAdmPath:      "./test/fixtures/discovery/cluster/fake_cibadmin.sh",
		CrmmonAdmPath:   "./test/fixtures/discovery/cluster/fake_crm_mon.sh",
		CorosyncKeyPath: "./test/fixtures/discovery/cluster/authkey",
		SBDPath:         "./test/fixtures/discovery/cluster/fake_sbd.sh",
		SBDConfigPath:   "./test/fixtures/discovery/cluster/sbd/sbd_config",
	})

	cluster.Provider = cloud.Azure

	return cluster
}
