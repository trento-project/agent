package crmmon

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/test/helpers"
)

type ParserTestSuite struct {
	suite.Suite
}

func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}

func (suite *ParserTestSuite) TestConstructor() {
	p := NewCrmMonParser("foo")
	suite.Equal("foo", p.crmMonPath)
}

func (suite *ParserTestSuite) TestParse() {
	p := NewCrmMonParser(helpers.GetFixturePath("discovery/cluster/fake_crm_mon.sh"))
	data, err := p.Parse()
	suite.NoError(err)
	suite.Equal("2.0.0", data.Version)
	suite.Equal(8, data.Summary.Resources.Number)
	suite.Equal(1, data.Summary.Resources.Disabled)
	suite.Equal(0, data.Summary.Resources.Blocked)
	suite.Equal("Fri Oct 18 11:48:22 2019", data.Summary.LastChange.Time)
	suite.Equal(2, data.Summary.Nodes.Number)
	suite.Equal("node01", data.Nodes[0].Name)
	suite.Equal("1084783375", data.Nodes[0].ID)
	suite.Equal(true, data.Nodes[0].Online)
	suite.Equal(true, data.Nodes[0].ExpectedUp)
	suite.Equal(true, data.Nodes[0].DC)
	suite.Equal(false, data.Nodes[0].Unclean)
	suite.Equal(false, data.Nodes[0].Shutdown)
	suite.Equal(false, data.Nodes[0].StandbyOnFail)
	suite.Equal(false, data.Nodes[0].Maintenance)
	suite.Equal(false, data.Nodes[0].Pending)
	suite.Equal(false, data.Nodes[0].Standby)
	suite.Equal("node02", data.Nodes[1].Name)
	suite.Equal("1084783376", data.Nodes[1].ID)
	suite.Equal(true, data.Nodes[1].Online)
	suite.Equal(true, data.Nodes[1].ExpectedUp)
	suite.Equal(false, data.Nodes[1].DC)
	suite.Equal(false, data.Nodes[1].Unclean)
	suite.Equal(false, data.Nodes[1].Shutdown)
	suite.Equal(false, data.Nodes[1].StandbyOnFail)
	suite.Equal(false, data.Nodes[1].Maintenance)
	suite.Equal(false, data.Nodes[1].Pending)
	suite.Equal(false, data.Nodes[1].Standby)
	suite.Equal("node01", data.NodeHistory.Nodes[0].Name)
	suite.Equal(5000, data.NodeHistory.Nodes[0].ResourceHistory[0].MigrationThreshold)
	suite.Equal(2, data.NodeHistory.Nodes[0].ResourceHistory[1].FailCount)
	suite.Equal("rsc_SAPHana_PRD_HDB00", data.NodeHistory.Nodes[0].ResourceHistory[0].Name)
	suite.Equal(4, len(data.Resources))
	suite.Equal("test-stop", data.Resources[0].ID)
	suite.Equal(false, data.Resources[0].Active)
	suite.Equal("Stopped", data.Resources[0].Role)
}

func (suite *ParserTestSuite) TestParseClones() {
	p := NewCrmMonParser(helpers.GetFixturePath("discovery/cluster/fake_crm_mon.sh"))
	data, err := p.Parse()
	suite.NoError(err)
	suite.Equal(3, len(data.Clones))
	suite.Equal("msl_SAPHana_PRD_HDB00", data.Clones[0].ID)
	suite.Equal("cln_SAPHanaTopology_PRD_HDB00", data.Clones[1].ID)
	suite.Equal("c-clusterfs", data.Clones[2].ID)
	suite.Equal(2, len(data.Clones[0].Resources))
	suite.Equal(2, len(data.Clones[1].Resources))
	suite.Equal("rsc_SAPHana_PRD_HDB00", data.Clones[0].Resources[0].ID)
	suite.Equal("Master", data.Clones[0].Resources[0].Role)
	suite.Equal("rsc_SAPHana_PRD_HDB00", data.Clones[0].Resources[1].ID)
	suite.Equal("Slave", data.Clones[0].Resources[1].Role)
}

func (suite *ParserTestSuite) TestParseGroups() {
	p := NewCrmMonParser(helpers.GetFixturePath("discovery/cluster/fake_crm_mon.sh"))
	data, err := p.Parse()
	suite.NoError(err)
	suite.Equal(2, len(data.Groups))

	suite.Equal("grp_HA1_ASCS00", data.Groups[0].ID)
	suite.Equal(3, len(data.Groups[0].Resources))
	suite.Equal("rsc_ip_HA1_ASCS00", data.Groups[0].Resources[0].ID)
	suite.Equal("rsc_fs_HA1_ASCS00", data.Groups[0].Resources[1].ID)
	suite.Equal("rsc_sap_HA1_ASCS00", data.Groups[0].Resources[2].ID)

	suite.Equal("grp_HA1_ERS10", data.Groups[1].ID)
	suite.Equal(3, len(data.Groups[1].Resources))
	suite.Equal("rsc_ip_HA1_ERS10", data.Groups[1].Resources[0].ID)
	suite.Equal("rsc_fs_HA1_ERS10", data.Groups[1].Resources[1].ID)
	suite.Equal("rsc_sap_HA1_ERS10", data.Groups[1].Resources[2].ID)
}

func (suite *ParserTestSuite) TestParseNodeAttributes() {
	p := NewCrmMonParser(helpers.GetFixturePath("discovery/cluster/fake_crm_mon.sh"))
	data, err := p.Parse()
	suite.NoError(err)
	suite.Len(data.NodeAttributes.Nodes, 2)
	suite.Equal("node01", data.NodeAttributes.Nodes[0].Name)
	suite.Equal("node02", data.NodeAttributes.Nodes[1].Name)

	suite.Len(data.NodeAttributes.Nodes[0].Attributes, 11)
	suite.Equal("hana_prd_clone_state", data.NodeAttributes.Nodes[0].Attributes[0].Name)
	suite.Equal("hana_prd_op_mode", data.NodeAttributes.Nodes[0].Attributes[1].Name)
	suite.Equal("hana_prd_remoteHost", data.NodeAttributes.Nodes[0].Attributes[2].Name)
	suite.Equal("hana_prd_roles", data.NodeAttributes.Nodes[0].Attributes[3].Name)
	suite.Equal("hana_prd_site", data.NodeAttributes.Nodes[0].Attributes[4].Name)
	suite.Equal("hana_prd_srmode", data.NodeAttributes.Nodes[0].Attributes[5].Name)
	suite.Equal("hana_prd_sync_state", data.NodeAttributes.Nodes[0].Attributes[6].Name)
	suite.Equal("hana_prd_version", data.NodeAttributes.Nodes[0].Attributes[7].Name)
	suite.Equal("hana_prd_vhost", data.NodeAttributes.Nodes[0].Attributes[8].Name)
	suite.Equal("lpa_prd_lpt", data.NodeAttributes.Nodes[0].Attributes[9].Name)
	suite.Equal("master-rsc_SAPHana_PRD_HDB00", data.NodeAttributes.Nodes[0].Attributes[10].Name)

	suite.Equal("PROMOTED", data.NodeAttributes.Nodes[0].Attributes[0].Value)
	suite.Equal("logreplay", data.NodeAttributes.Nodes[0].Attributes[1].Value)
	suite.Equal("node02", data.NodeAttributes.Nodes[0].Attributes[2].Value)
	suite.Equal("4:P:master1:master:worker:master", data.NodeAttributes.Nodes[0].Attributes[3].Value)
	suite.Equal("PRIMARY_SITE_NAME", data.NodeAttributes.Nodes[0].Attributes[4].Value)
	suite.Equal("sync", data.NodeAttributes.Nodes[0].Attributes[5].Value)
	suite.Equal("PRIM", data.NodeAttributes.Nodes[0].Attributes[6].Value)
	suite.Equal("2.00.040.00.1553674765", data.NodeAttributes.Nodes[0].Attributes[7].Value)
	suite.Equal("node01", data.NodeAttributes.Nodes[0].Attributes[8].Value)
	suite.Equal("1571392102", data.NodeAttributes.Nodes[0].Attributes[9].Value)
	suite.Equal("150", data.NodeAttributes.Nodes[0].Attributes[10].Value)

	suite.Len(data.NodeAttributes.Nodes[1].Attributes, 11)
	suite.Equal("hana_prd_clone_state", data.NodeAttributes.Nodes[0].Attributes[0].Name)
	suite.Equal("hana_prd_op_mode", data.NodeAttributes.Nodes[0].Attributes[1].Name)
	suite.Equal("hana_prd_remoteHost", data.NodeAttributes.Nodes[0].Attributes[2].Name)
	suite.Equal("hana_prd_roles", data.NodeAttributes.Nodes[0].Attributes[3].Name)
	suite.Equal("hana_prd_site", data.NodeAttributes.Nodes[0].Attributes[4].Name)
	suite.Equal("hana_prd_srmode", data.NodeAttributes.Nodes[0].Attributes[5].Name)
	suite.Equal("hana_prd_sync_state", data.NodeAttributes.Nodes[0].Attributes[6].Name)
	suite.Equal("hana_prd_version", data.NodeAttributes.Nodes[0].Attributes[7].Name)
	suite.Equal("hana_prd_vhost", data.NodeAttributes.Nodes[0].Attributes[8].Name)
	suite.Equal("lpa_prd_lpt", data.NodeAttributes.Nodes[0].Attributes[9].Name)
	suite.Equal("master-rsc_SAPHana_PRD_HDB00", data.NodeAttributes.Nodes[0].Attributes[10].Name)

	suite.Equal("DEMOTED", data.NodeAttributes.Nodes[1].Attributes[0].Value)
	suite.Equal("logreplay", data.NodeAttributes.Nodes[1].Attributes[1].Value)
	suite.Equal("node01", data.NodeAttributes.Nodes[1].Attributes[2].Value)
	suite.Equal("4:S:master1:master:worker:master", data.NodeAttributes.Nodes[1].Attributes[3].Value)
	suite.Equal("SECONDARY_SITE_NAME", data.NodeAttributes.Nodes[1].Attributes[4].Value)
	suite.Equal("sync", data.NodeAttributes.Nodes[1].Attributes[5].Value)
	suite.Equal("SOK", data.NodeAttributes.Nodes[1].Attributes[6].Value)
	suite.Equal("2.00.040.00.1553674765", data.NodeAttributes.Nodes[1].Attributes[7].Value)
	suite.Equal("node02", data.NodeAttributes.Nodes[1].Attributes[8].Value)
	suite.Equal("30", data.NodeAttributes.Nodes[1].Attributes[9].Value)
	suite.Equal("100", data.NodeAttributes.Nodes[1].Attributes[10].Value)
}
