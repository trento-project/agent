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

	factsRequest := []entities.FactRequest{
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

	factsGathered, err := c.Gather(factsRequest)

	expectedResults := []entities.Fact{
		{
			Name:  "corosync_token",
			Value: "30000",
			Error: nil,
		},
		{
			Name:  "corosync_join",
			Value: "60",
			Error: nil,
		},
		{
			Name:  "corosync_node1id",
			Value: "1",
			Error: nil,
		},
		{
			Name:  "corosync_node2id",
			Value: "2",
			Error: nil,
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
			Error: nil,
		},
		{
			Name:  "corosync_not_found",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "requested field value not found: totem.not_found",
				Type:    "corosync-conf-value-not-found",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factsGathered)
}

func (suite *CorosyncConfTestSuite) TestCorosyncConfFileNotExists() {
	c := NewCorosyncConfGatherer("not_found")

	factsRequest := []entities.FactRequest{
		{
			Name:     "corosync_token",
			Gatherer: "corosync.conf",
			Argument: "totem.token",
		},
	}

	factsGathered, err := c.Gather(factsRequest)

	expectedError := &entities.FactGatheringError{
		Message: "error reading corosync.conf file: could not open corosync.conf file: " +
			"open not_found: no such file or directory",
		Type: "corosync-conf-file-error",
	}

	suite.EqualError(err, expectedError.Error())
	suite.Empty(factsGathered)
}

func (suite *CorosyncConfTestSuite) TestCorosyncConfInvalid() {
	c := NewCorosyncConfGatherer("../../../test/fixtures/gatherers/corosync.conf.invalid")

	factsRequest := []entities.FactRequest{
		{
			Name:     "corosync_token",
			Gatherer: "corosync.conf",
			Argument: "totem.token",
		},
	}

	factsGathered, err := c.Gather(factsRequest)

	expectedError := &entities.FactGatheringError{
		Message: "error decoding corosync.conf file: invalid corosync file structure. " +
			"some section is not closed properly",
		Type: "corosync-conf-decoding-error",
	}

	suite.EqualError(err, expectedError.Error())
	suite.Empty(factsGathered)
}
