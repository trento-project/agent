package gatherers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/test/helpers"
)

type HostsFileTestSuite struct {
	suite.Suite
}

func TestHostsFileTestSuite(t *testing.T) {
	suite.Run(t, new(HostsFileTestSuite))
}

func (suite *HostsFileTestSuite) TestHostsFileBasic() {
	c := gatherers.NewHostsFileGatherer(helpers.GetFixturePath("gatherers/hosts.basic"))

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
		{
			Name:     "hosts_all",
			Gatherer: "hosts",
			CheckID:  "check4",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

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
		{
			Name: "hosts_all",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"localhost": &entities.FactValueList{Value: []entities.FactValue{
						&entities.FactValueString{Value: "127.0.0.1"},
						&entities.FactValueString{Value: "::1"},
					}},
					"somehost": &entities.FactValueList{Value: []entities.FactValue{
						&entities.FactValueString{Value: "127.0.1.1"},
					}},
					"suse.com": &entities.FactValueList{Value: []entities.FactValue{
						&entities.FactValueString{Value: "52.84.66.74"},
					}},
					"ip6-localhost": &entities.FactValueList{Value: []entities.FactValue{
						&entities.FactValueString{Value: "::1"},
					}},
					"ip6-loopback": &entities.FactValueList{Value: []entities.FactValue{
						&entities.FactValueString{Value: "::1"},
					}},
					"ip6-allnodes": &entities.FactValueList{Value: []entities.FactValue{
						&entities.FactValueString{Value: "ff02::1"},
					}},
					"ip6-allrouters": &entities.FactValueList{Value: []entities.FactValue{
						&entities.FactValueString{Value: "ff02::2"},
					}},
				},
			},
			CheckID: "check4",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *HostsFileTestSuite) TestHostsFileNotExists() {
	c := gatherers.NewHostsFileGatherer("non_existing_file")

	factRequests := []entities.FactRequest{
		{
			Name:     "hosts_somehost",
			Gatherer: "hosts",
			Argument: "somehost",
		},
	}

	_, err := c.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: hosts-file-error - error reading /etc/hosts file: "+
		"open non_existing_file: no such file or directory")
}

func (suite *HostsFileTestSuite) TestHostsFileIgnoresCommentedHosts() {

	c := gatherers.NewHostsFileGatherer(helpers.GetFixturePath("gatherers/hosts.basic"))

	factRequests := []entities.FactRequest{
		{
			Name:     "hosts_commented-host",
			Gatherer: "hosts",
			Argument: "commented-host",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

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
