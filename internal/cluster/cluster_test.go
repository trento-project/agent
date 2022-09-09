//nolint:exhaustruct
package cluster

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/cluster/cib"
	"github.com/trento-project/agent/internal/cluster/crmmon"
)

type ClusterTestSuite struct {
	suite.Suite
}

func TestClusterTestSuite(t *testing.T) {
	suite.Run(t, new(ClusterTestSuite))
}

func (suite *ClusterTestSuite) TestClusterId() {
	root := new(cib.Root)

	c := Cluster{
		Cib:  *root,
		Name: "sculpin",
		ID:   "47d1190ffb4f781974c8356d7f863b03",
	}

	authkey, _ := getCorosyncAuthkeyMd5("../../test/authkey")

	suite.Equal(c.ID, authkey)
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

	c := Cluster{
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

	c := &Cluster{
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

	suite.Equal(true, isDC(c))

	c = &Cluster{
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

	suite.Equal(false, isDC(c))
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

	c := Cluster{
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

	c = Cluster{
		Cib: *root,
	}

	suite.Equal(false, c.IsFencingEnabled())
}

func (suite *ClusterTestSuite) TestFencingType() {
	c := Cluster{
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

	c = Cluster{
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
	c := Cluster{
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

	c = Cluster{
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
	c := Cluster{
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

	c = Cluster{
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
