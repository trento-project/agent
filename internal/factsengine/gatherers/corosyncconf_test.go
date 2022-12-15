package gatherers

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/test/helpers"
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
	c := NewCorosyncConfGatherer(helpers.GetFixturePath("gatherers/corosync.conf.basic"))

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
			Name:     "corosync_interfaces",
			Gatherer: "corosync.conf",
			Argument: "totem.interface",
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
			Value: &entities.FactValueInt{Value: 30000},
			Error: nil,
		},
		{
			Name:  "corosync_join",
			Value: &entities.FactValueInt{Value: 60},
			Error: nil,
		},
		{
			Name: "corosync_interfaces",
			Value: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"ringnumber":  &entities.FactValueInt{Value: 0},
					"bindnetaddr": &entities.FactValueString{Value: "192.168.1.0"},
					"mcastport":   &entities.FactValueInt{Value: 5405},
					"ttl":         &entities.FactValueInt{Value: 1},
				}},
			}},
			Error: nil,
		},
		{
			Name:  "corosync_node1id",
			Value: &entities.FactValueInt{Value: 1},
			Error: nil,
		},
		{
			Name:  "corosync_node2id",
			Value: &entities.FactValueInt{Value: 2},
			Error: nil,
		},
		{
			Name: "corosync_nodes",
			Value: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"ring0_addr": &entities.FactValueString{Value: "10.0.0.119"},
					"ring1_addr": &entities.FactValueString{Value: "10.0.0.120"},
					"nodeid":     &entities.FactValueInt{Value: 1},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"ring0_addr": &entities.FactValueString{Value: "10.0.1.153"},
					"ring1_addr": &entities.FactValueString{Value: "10.0.1.154"},
					"nodeid":     &entities.FactValueInt{Value: 2},
				}},
			}},
			Error: nil,
		},
		{
			Name:  "corosync_not_found",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error getting value: requested field value not found: totem.not_found",
				Type:    "value-not-found",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factsGathered)
}

func (suite *CorosyncConfTestSuite) TestCorosyncConfOneNode() {
	c := NewCorosyncConfGatherer(helpers.GetFixturePath("gatherers/corosync.conf.one_node"))

	factsRequest := []entities.FactRequest{
		{
			Name:     "corosync_nodes",
			Gatherer: "corosync.conf",
			Argument: "nodelist.node",
		},
	}

	factsGathered, err := c.Gather(factsRequest)

	expectedResults := []entities.Fact{

		{
			Name: "corosync_nodes",
			Value: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"ring0_addr": &entities.FactValueString{Value: "10.0.0.119"},
					"ring1_addr": &entities.FactValueString{Value: "10.0.0.120"},
					"nodeid":     &entities.FactValueInt{Value: 1},
				}},
			}},
			Error: nil,
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factsGathered)
}

func (suite *CorosyncConfTestSuite) TestCorosyncConfThreeNodes() {
	c := NewCorosyncConfGatherer(helpers.GetFixturePath("gatherers/corosync.conf.three_node"))

	factsRequest := []entities.FactRequest{
		{
			Name:     "corosync_nodes",
			Gatherer: "corosync.conf",
			Argument: "nodelist.node",
		},
	}

	factsGathered, err := c.Gather(factsRequest)

	expectedResults := []entities.Fact{

		{
			Name: "corosync_nodes",
			Value: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"ring0_addr": &entities.FactValueString{Value: "10.0.0.119"},
					"ring1_addr": &entities.FactValueString{Value: "10.0.0.120"},
					"nodeid":     &entities.FactValueInt{Value: 1},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"ring0_addr": &entities.FactValueString{Value: "10.0.1.153"},
					"ring1_addr": &entities.FactValueString{Value: "10.0.1.154"},
					"nodeid":     &entities.FactValueInt{Value: 2},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"ring0_addr": &entities.FactValueString{Value: "10.0.1.155"},
					"ring1_addr": &entities.FactValueString{Value: "10.0.1.156"},
					"nodeid":     &entities.FactValueInt{Value: 3},
				}},
			}},
			Error: nil,
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
	c := NewCorosyncConfGatherer(helpers.GetFixturePath("gatherers/corosync.conf.invalid"))

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
