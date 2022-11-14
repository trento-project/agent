package gatherers

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/agent/test/helpers"
)

type HostsFileTestSuite struct {
	suite.Suite
}

func TestHostsFileTestSuite(t *testing.T) {
	suite.Run(t, new(HostsFileTestSuite))
}

func (suite *HostsFileTestSuite) TestHostsFileBasic() {
	c := NewHostsFileGatherer(helpers.GetFixturePath("gatherers/hosts.basic"))

	factRequests := []entities.FactRequest{
		{
			Name:     "hosts_localhost",
			Gatherer: "hosts",
			Argument: "localhost",
			CheckID:  "check1",
		},
		{
			Name:     "hosts_somehost",
			Gatherer: "hosts",
			Argument: "somehost",
			CheckID:  "check2",
		},
		{
			Name:     "hosts_ip6-localhost",
			Gatherer: "hosts",
			Argument: "ip6-localhost",
			CheckID:  "check3",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "hosts_localhost",
			Value: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueString{Value: "127.0.0.1"},
				&entities.FactValueString{Value: "::1"},
			}},
			CheckID: "check1",
		},
		{
			Name: "hosts_somehost",
			Value: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueString{Value: "127.0.1.1"},
			}},
			CheckID: "check2",
		},
		{
			Name: "hosts_ip6-localhost",
			Value: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueString{Value: "::1"},
			}},
			CheckID: "check3",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *HostsFileTestSuite) TestHostsFileNotExists() {
	c := NewHostsFileGatherer("non_existing_file")

	factRequests := []entities.FactRequest{
		{
			Name:     "hosts_somehost",
			Gatherer: "hosts",
			Argument: "somehost",
		},
	}

	_, err := c.Gather(factRequests)

	suite.EqualError(err, "fact gathering error: hosts-file-error - error reading /etc/hosts file: open non_existing_file: no such file or directory")
}

func (suite *HostsFileTestSuite) TestHostsFileIgnoresCommentedHosts() {

	c := NewHostsFileGatherer(helpers.GetFixturePath("gatherers/hosts.basic"))

	factRequests := []entities.FactRequest{
		{
			Name:     "hosts_commented-host",
			Gatherer: "hosts",
			Argument: "commented-host",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "hosts_commented-host",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "requested field value not found in /etc/hosts file: commented-host",
				Type:    "hosts-file-value-not-found",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
