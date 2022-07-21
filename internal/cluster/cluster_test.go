//nolint:exhaustruct
package cluster

import (
	"os"
	"testing"

	"github.com/trento-project/agent/internal/cluster/cib"
	"github.com/trento-project/agent/internal/cluster/crmmon"

	"github.com/stretchr/testify/assert"
)

func TestClusterId(t *testing.T) {
	root := new(cib.Root)

	c := Cluster{
		Cib:  *root,
		Name: "sculpin",
		ID:   "47d1190ffb4f781974c8356d7f863b03",
	}

	authkey, _ := getCorosyncAuthkeyMd5("../../test/authkey")

	assert.Equal(t, c.ID, authkey)
}

func TestClusterName(t *testing.T) {
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

	assert.Equal(t, "cluster_name", c.Name)
}

func TestIsDC(t *testing.T) {
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

	assert.Equal(t, true, isDC(c))

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

	assert.Equal(t, false, isDC(c))
}

func TestIsFencingEnabled(t *testing.T) {
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

	assert.Equal(t, true, c.IsFencingEnabled())

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

	assert.Equal(t, false, c.IsFencingEnabled())
}

func TestFencingType(t *testing.T) {
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

	assert.Equal(t, "myfencing", c.FencingType())

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

	assert.Equal(t, "notconfigured", c.FencingType())
}

func TestFencingResourceExists(t *testing.T) {
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

	assert.Equal(t, true, c.FencingResourceExists())

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

	assert.Equal(t, false, c.FencingResourceExists())
}

func TestIsFencingSBD(t *testing.T) {
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

	assert.Equal(t, true, c.IsFencingSBD())

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

	assert.Equal(t, false, c.IsFencingSBD())
}
