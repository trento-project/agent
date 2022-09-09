//nolint:lll
package cib

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ParserTestSuite struct {
	suite.Suite
}

func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}

func (suite *ParserTestSuite) TestConstructor() {
	p := NewCibAdminParser("foo")
	suite.Equal("foo", p.cibAdminPath)
}

func (suite *ParserTestSuite) TestParse() {
	p := NewCibAdminParser("../../../test/fake_cibadmin.sh")
	data, err := p.Parse()
	suite.NoError(err)
	suite.Equal(2, len(data.Configuration.Nodes))
	suite.Equal("cib-bootstrap-options-cluster-name", data.Configuration.CrmConfig.ClusterProperties[3].ID)
	suite.Equal("hana_cluster", data.Configuration.CrmConfig.ClusterProperties[3].Value)
	suite.Equal("node01", data.Configuration.Nodes[0].Uname)
	suite.Equal("node02", data.Configuration.Nodes[1].Uname)
	suite.Equal(4, len(data.Configuration.Resources.Primitives))
	suite.Equal(1, len(data.Configuration.Resources.Masters))
	suite.Equal(1, len(data.Configuration.Resources.Clones))
	suite.Equal("stonith-sbd", data.Configuration.Resources.Primitives[0].ID)
	suite.Equal("stonith", data.Configuration.Resources.Primitives[0].Class)
	suite.Equal("external/sbd", data.Configuration.Resources.Primitives[0].Type)
	suite.Equal(1, len(data.Configuration.Resources.Primitives[0].InstanceAttributes))
	suite.Equal("pcmk_delay_max", data.Configuration.Resources.Primitives[0].InstanceAttributes[0].Name)
	suite.Equal("stonith-sbd-instance_attributes-pcmk_delay_max", data.Configuration.Resources.Primitives[0].InstanceAttributes[0].ID)
	suite.Equal("30s", data.Configuration.Resources.Primitives[0].InstanceAttributes[0].Value)
	suite.Equal(1, len(data.Configuration.Resources.Groups))
	suite.Equal("g_ip_PRD_HDB00", data.Configuration.Resources.Groups[0].ID)
	suite.Equal(1, len(data.Configuration.Resources.Groups[0].Primitives))
	suite.Equal("rsc_ip_PRD_HDB00", data.Configuration.Resources.Groups[0].Primitives[0].ID)
	suite.Equal("ocf", data.Configuration.Resources.Groups[0].Primitives[0].Class)
	suite.Equal("heartbeat", data.Configuration.Resources.Groups[0].Primitives[0].Provider)
	suite.Equal("IPaddr2", data.Configuration.Resources.Groups[0].Primitives[0].Type)
	suite.Equal(1, len(data.Configuration.Resources.Groups[0].Primitives[0].InstanceAttributes))
	suite.Equal("rsc_ip_PRD_HDB00-instance_attributes-ip", data.Configuration.Resources.Groups[0].Primitives[0].InstanceAttributes[0].ID)
	suite.Equal("ip", data.Configuration.Resources.Groups[0].Primitives[0].InstanceAttributes[0].Name)
	suite.Equal("10.74.1.12", data.Configuration.Resources.Groups[0].Primitives[0].InstanceAttributes[0].Value)
	suite.Equal("msl_SAPHana_PRD_HDB00", data.Configuration.Resources.Masters[0].ID)
	suite.Equal(3, len(data.Configuration.Resources.Masters[0].MetaAttributes))
	suite.Equal("rsc_SAPHana_PRD_HDB00", data.Configuration.Resources.Masters[0].Primitive.ID)
	suite.Equal(5, len(data.Configuration.Resources.Masters[0].Primitive.Operations))
	suite.Equal("rsc_SAPHana_PRD_HDB00-start-0", data.Configuration.Resources.Masters[0].Primitive.Operations[0].ID)
	suite.Equal("start", data.Configuration.Resources.Masters[0].Primitive.Operations[0].Name)
	suite.Equal("0", data.Configuration.Resources.Masters[0].Primitive.Operations[0].Interval)
	suite.Equal("3600", data.Configuration.Resources.Masters[0].Primitive.Operations[0].Timeout)
	suite.Equal("rsc_SAPHana_PRD_HDB00-stop-0", data.Configuration.Resources.Masters[0].Primitive.Operations[1].ID)
	suite.Equal("stop", data.Configuration.Resources.Masters[0].Primitive.Operations[1].Name)
	suite.Equal("0", data.Configuration.Resources.Masters[0].Primitive.Operations[1].Interval)
	suite.Equal("3600", data.Configuration.Resources.Masters[0].Primitive.Operations[1].Timeout)
	suite.Equal("rsc_SAPHana_PRD_HDB00-promote-0", data.Configuration.Resources.Masters[0].Primitive.Operations[2].ID)
	suite.Equal("promote", data.Configuration.Resources.Masters[0].Primitive.Operations[2].Name)
	suite.Equal("0", data.Configuration.Resources.Masters[0].Primitive.Operations[2].Interval)
	suite.Equal("3600", data.Configuration.Resources.Masters[0].Primitive.Operations[2].Timeout)
	suite.Equal("rsc_SAPHana_PRD_HDB00-monitor-60", data.Configuration.Resources.Masters[0].Primitive.Operations[3].ID)
	suite.Equal("monitor", data.Configuration.Resources.Masters[0].Primitive.Operations[3].Name)
	suite.Equal("Master", data.Configuration.Resources.Masters[0].Primitive.Operations[3].Role)
	suite.Equal("60", data.Configuration.Resources.Masters[0].Primitive.Operations[3].Interval)
	suite.Equal("700", data.Configuration.Resources.Masters[0].Primitive.Operations[3].Timeout)
	suite.Equal("rsc_SAPHana_PRD_HDB00-monitor-61", data.Configuration.Resources.Masters[0].Primitive.Operations[4].ID)
	suite.Equal("monitor", data.Configuration.Resources.Masters[0].Primitive.Operations[4].Name)
	suite.Equal("Slave", data.Configuration.Resources.Masters[0].Primitive.Operations[4].Role)
	suite.Equal("61", data.Configuration.Resources.Masters[0].Primitive.Operations[4].Interval)
	suite.Equal("700", data.Configuration.Resources.Masters[0].Primitive.Operations[4].Timeout)
	suite.Equal("test", data.Configuration.Resources.Primitives[2].ID)
	suite.Equal("ocf", data.Configuration.Resources.Primitives[2].Class)
	suite.Equal("heartbeat", data.Configuration.Resources.Primitives[2].Provider)
	suite.Equal("Dummy", data.Configuration.Resources.Primitives[2].Type)
}
