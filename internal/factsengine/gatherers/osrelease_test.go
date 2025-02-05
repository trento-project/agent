package gatherers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/test/helpers"
)

type OSReleaseGathererTestSuite struct {
	suite.Suite
}

func TestOSReleaseGathererTestSuite(t *testing.T) {
	suite.Run(t, new(OSReleaseGathererTestSuite))
}

func (suite *OSReleaseGathererTestSuite) TestOSReleaseGathererSuccess() {
	c := gatherers.NewOSReleaseGatherer(helpers.GetFixturePath("gatherers/os-release.basic"))

	factRequests := []entities.FactRequest{
		{
			Name:     "os-release",
			Gatherer: "os-release@v1",
			CheckID:  "check1",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "os-release",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"NAME":           &entities.FactValueString{Value: "openSUSE Leap"},
					"VERSION":        &entities.FactValueString{Value: "15.2"},
					"ID":             &entities.FactValueString{Value: "opensuse-leap"},
					"ID_LIKE":        &entities.FactValueString{Value: "suse opensuse"},
					"VERSION_ID":     &entities.FactValueString{Value: "15.2"},
					"PRETTY_NAME":    &entities.FactValueString{Value: "openSUSE Leap 15.2"},
					"ANSI_COLOR":     &entities.FactValueString{Value: "0;32"},
					"CPE_NAME":       &entities.FactValueString{Value: "cpe:/o:opensuse:leap:15.2"},
					"BUG_REPORT_URL": &entities.FactValueString{Value: "https://bugs.opensuse.org"},
					"HOME_URL":       &entities.FactValueString{Value: "https://www.opensuse.org/"},
				},
			},
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *OSReleaseGathererTestSuite) TestOSReleaseGathererFileNotExists() {
	c := gatherers.NewOSReleaseGatherer("non_existing_file")

	factRequests := []entities.FactRequest{
		{
			Name:     "os-release",
			Gatherer: "os-release@v1",
		},
	}

	_, err := c.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: os-release-file-error - error reading /etc/os-release file: "+
		"open non_existing_file: no such file or directory")
}

func (suite *OSReleaseGathererTestSuite) TestOSReleaseGathererErrorDecoding() {

	c := gatherers.NewOSReleaseGatherer(helpers.GetFixturePath("gatherers/os-release.invalid"))

	factRequests := []entities.FactRequest{
		{
			Name:     "os-release",
			Gatherer: "os-release",
		},
	}

	_, err := c.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: os-release-decoding-error - error decoding file content: error on line 3: missing =")
}

func (suite *OSReleaseGathererTestSuite) TestOSReleaseContextCancelled() {
	gatherer := gatherers.NewOSReleaseGatherer(helpers.GetFixturePath("gatherers/os-release.basic"))

	factsRequest := []entities.FactRequest{{
		Name:     "os-release",
		Gatherer: "os-release@v1",
		CheckID:  "check1",
	}}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	factResults, err := gatherer.Gather(ctx, factsRequest)

	suite.Error(err)
	suite.Empty(factResults)
}
