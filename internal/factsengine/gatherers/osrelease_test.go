package gatherers_test

import (
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

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "os-release",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"NAME":               &entities.FactValueString{Value: "Ubuntu"},
					"VERSION":            &entities.FactValueString{Value: "20.04.3 LTS (Focal Fossa)"},
					"ID":                 &entities.FactValueString{Value: "ubuntu"},
					"ID_LIKE":            &entities.FactValueString{Value: "debian"},
					"VERSION_ID":         &entities.FactValueString{Value: "20.04"},
					"PRETTY_NAME":        &entities.FactValueString{Value: "Ubuntu 20.04.3 LTS"},
					"VERSION_CODENAME":   &entities.FactValueString{Value: "focal"},
					"HOME_URL":           &entities.FactValueString{Value: "https://www.ubuntu.com/"},
					"SUPPORT_URL":        &entities.FactValueString{Value: "https://help.ubuntu.com/"},
					"BUG_REPORT_URL":     &entities.FactValueString{Value: "https://bugs.launchpad.net/ubuntu/"},
					"PRIVACY_POLICY_URL": &entities.FactValueString{Value: "https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"},
					"UBUNTU_CODENAME":    &entities.FactValueString{Value: "focal"},
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

	_, err := c.Gather(factRequests)

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

	_, err := c.Gather(factRequests)

	suite.EqualError(err, "fact gathering error: os-release-decoding-error - error decoding file content: error on line 3: missing =")
}
