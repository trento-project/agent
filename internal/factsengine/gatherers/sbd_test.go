package gatherers_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

type SBDGathererTestSuite struct {
	suite.Suite
	configurationFile string
}

func TestSBDGathererTestSuite(t *testing.T) {
	sbdSuite := new(SBDGathererTestSuite)
	sbdSuite.configurationFile = "../../../test/sbd_config"
	suite.Run(t, sbdSuite)
}

func (suite *SBDGathererTestSuite) TestConfigFileCouldNotBeRead() {
	const testSBDConfig = "../../../test/some-non-existent-sbd-config"

	requestedFacts := []entities.FactRequest{
		{
			Name:     "sbd_pacemaker",
			Gatherer: "sbd_config",
			Argument: "SBD_PACEMAKER",
		},
		{
			Name:     "sbd_unexistent",
			Gatherer: "sbd_config",
			Argument: "SBD_THIS_DOES_NOT_EXIST",
		},
	}

	gatherer := gatherers.NewSBDGatherer(testSBDConfig)

	gatheredFacts, err := gatherer.Gather(requestedFacts)

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:  "sbd_pacemaker",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error parsing SBD configuration file: " +
					"open ../../../test/some-non-existent-sbd-config: no such file or directory",
				Type: "sbd-config-parsing-error",
			},
		},
		{
			Name:  "sbd_unexistent",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error parsing SBD configuration file: " +
					"open ../../../test/some-non-existent-sbd-config: no such file or directory",
				Type: "sbd-config-parsing-error",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, gatheredFacts)
}

func (suite *SBDGathererTestSuite) TestSomeRequiredValueDoesNotExistInConfig() {
	requestedFacts := []entities.FactRequest{
		{
			Name:     "sbd_pacemaker",
			Gatherer: "sbd_config",
			Argument: "SBD_PACEMAKER",
		},
		{
			Name:     "sbd_unexistent",
			Gatherer: "sbd_config",
			Argument: "SBD_THIS_DOES_NOT_EXIST",
		},
	}

	gatherer := gatherers.NewSBDGatherer(suite.configurationFile)

	gatheredFacts, err := gatherer.Gather(requestedFacts)

	expectedFacts := []entities.FactsGatheredItem{
		{
			Name:  "sbd_pacemaker",
			Value: "yes",
			Error: nil,
		},
		{
			Name:  "sbd_unexistent",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "requested field value not found: requested value SBD_THIS_DOES_NOT_EXIST not found",
				Type:    "sbd-config-value-not-found",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedFacts, gatheredFacts)
}

func (suite *SBDGathererTestSuite) TestSBDGatherer() {
	requestedFacts := []entities.FactRequest{
		{
			Name:     "sbd_pacemaker",
			Gatherer: "sbd_config",
			Argument: "SBD_PACEMAKER",
		},
		{
			Name:     "sbd_startmode",
			Gatherer: "sbd_config",
			Argument: "SBD_STARTMODE",
		},
		{
			Name:     "sbd_test_value",
			Gatherer: "sbd_config",
			Argument: "TEST2",
		},
	}

	gatherer := gatherers.NewSBDGatherer(suite.configurationFile)

	gatheredFacts, _ := gatherer.Gather(requestedFacts)

	expectedFacts := []entities.FactsGatheredItem{
		{
			Name:  "sbd_pacemaker",
			Value: "yes",
		},
		{
			Name:  "sbd_startmode",
			Value: "always",
		},
		{
			Name:  "sbd_test_value",
			Value: "Value2",
		},
	}

	suite.ElementsMatch(expectedFacts, gatheredFacts)
}
