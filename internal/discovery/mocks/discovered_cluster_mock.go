package mocks

import (
	"github.com/trento-project/agent/internal/core/cloud"
	"github.com/trento-project/agent/internal/core/cluster"
	mocksUtils "github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

func mockSbdDump() []byte {
	output := `==Dumping header on disk /dev/vdb
Header version     : 2.1
UUID               : f9ba490e-0f14-4908-859a-ace97aafaf34
Number of slots    : 255
Sector size        : 512
Timeout (watchdog) : 5
Timeout (allocate) : 2
Timeout (loop)     : 1
Timeout (msgwait)  : 10
==Header on disk /dev/vdb is dumped`
	return []byte(output)
}

func mockSbdList() []byte {
	output := `0	vmhana01	clear
1	vmhana02	clear`
	return []byte(output)
}

func NewDiscoveredClusterMock() *cluster.Cluster {
	mockCommand := new(mocksUtils.CommandExecutor)
	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").
		Return([]byte("7783-7084-3265-9085-8269-3286-77"), nil)
	mockCommand.On("Exec", "/usr/sbin/sbd", "-d", "/dev/vdb", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Exec", "/usr/sbin/sbd", "-d", "/dev/vdb", "list").Return(mockSbdList(), nil)
	mockCommand.On("Exec", "/usr/sbin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Exec", "/usr/sbin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdList(), nil)

	cluster, _ := cluster.NewClusterWithDiscoveryTools(&cluster.DiscoveryTools{
		CibAdmPath:      helpers.GetFixturePath("discovery/cluster/fake_cibadmin.sh"),
		CrmmonAdmPath:   helpers.GetFixturePath("discovery/cluster/fake_crm_mon.sh"),
		CorosyncKeyPath: helpers.GetFixturePath("discovery/cluster/authkey"),
		SBDPath:         "/usr/sbin/sbd",
		SBDConfigPath:   helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"),
		CommandExecutor: mockCommand,
	})

	cluster.Provider = cloud.Azure

	return cluster
}
