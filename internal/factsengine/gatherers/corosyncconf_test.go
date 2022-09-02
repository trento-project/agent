package gatherers

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

type CorosyncConfTestSuite struct {
	suite.Suite
}

func TestCorosyncConfTestSuite(t *testing.T) {
	suite.Run(t, new(CorosyncConfTestSuite))
}

func (suite *CorosyncConfTestSuite) TestCorosyncConfDefault() {
	c := NewDefaultCorosyncConfGatherer()
	suite.Equal("/etc/corosync/corosync.conf", c.configFile)
}

func (suite *CorosyncConfTestSuite) TestCorosyncConfBasic() {
	c := NewCorosyncConfGatherer("../../../test/fixtures/gatherers/corosync.conf.basic")

	factRequests := []entities.FactRequest{
		{
			Name:     "corosync_token",
			Gatherer: "corosync.conf",
			Argument: "totem.token",
		},
		{
			Name:     "corosync_join",
			Gatherer: "corosync.conf",
			Argument: "totem.join",
		},
		{
			Name:     "corosync_node1id",
			Gatherer: "corosync.conf",
			Argument: "nodelist.node.0.nodeid",
		},
		{
			Name:     "corosync_node2id",
			Gatherer: "corosync.conf",
			Argument: "nodelist.node.1.nodeid",
		},
		{
			Name:     "corosync_nodes",
			Gatherer: "corosync.conf",
			Argument: "nodelist.node",
		},
		{
			Name:     "corosync_not_found",
			Gatherer: "corosync.conf",
			Argument: "totem.not_found",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:  "corosync_token",
			Value: "30000",
		},
		{
			Name:  "corosync_join",
			Value: "60",
		},
		{
			Name:  "corosync_node1id",
			Value: "1",
		},
		{
			Name:  "corosync_node2id",
			Value: "2",
		},
		{
			Name: "corosync_nodes",
			Value: []interface{}{
				map[string]interface{}{
					"ring0_addr": "10.0.0.119",
					"ring1_addr": "10.0.0.120",
					"nodeid":     "1",
				},
				map[string]interface{}{
					"ring0_addr": "10.0.1.153",
					"ring1_addr": "10.0.1.154",
					"nodeid":     "2",
				},
			},
		},
		{
			Name:  "corosync_not_found",
			Value: nil,
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *CorosyncConfTestSuite) TestCorosyncConfFileNotExists() {
	c := NewCorosyncConfGatherer("not_found")

	factRequests := []entities.FactRequest{
		{
			Name:     "corosync_token",
			Gatherer: "corosync.conf",
			Argument: "totem.token",
		},
	}

	_, err := c.Gather(factRequests)

	suite.EqualError(err, "could not open corosync.conf file: open not_found: no such file or directory")
}

func (suite *CorosyncConfTestSuite) TestCorosyncConfInvalid() {
	c := NewCorosyncConfGatherer("../../../test/fixtures/gatherers/corosync.conf.invalid")

	factRequests := []entities.FactRequest{
		{
			Name:     "corosync_token",
			Gatherer: "corosync.conf",
			Argument: "totem.token",
		},
	}

	_, err := c.Gather(factRequests)

	suite.EqualError(err, "invalid corosync file structure. some section is not closed properly")
}
