package gatherers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/test/helpers"
)

type SBDGathererTestSuite struct {
	suite.Suite
}

func TestSBDGathererTestSuite(t *testing.T) {
	suite.Run(t, new(SBDGathererTestSuite))
}

func (suite *SBDGathererTestSuite) TestConfigFileNoArgumentProvided() {
	requestedFacts := []entities.FactRequest{
		{
			Name:     "no_argument_fact",
			Gatherer: "sbd_config",
		},
		{
			Name:     "empty_argument_fact",
			Gatherer: "sbd_config",
			Argument: "",
		},
	}

	gatherer := gatherers.NewSBDGatherer(helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"))

	gatheredFacts, err := gatherer.Gather(context.Background(), requestedFacts)

	expectedFacts := []entities.Fact{
		{
			Name:  "no_argument_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Type:    "sbd-config-missing-argument",
				Message: "missing required argument",
			},
		},
		{
			Name:  "empty_argument_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Type:    "sbd-config-missing-argument",
				Message: "missing required argument",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedFacts, gatheredFacts)
}

func (suite *SBDGathererTestSuite) TestConfigFileCouldNotBeRead() {
	requestedFacts := []entities.FactRequest{}

	gatherer := gatherers.NewSBDGatherer("/path/to/some-non-existent-sbd-config")

	gatheredFacts, err := gatherer.Gather(context.Background(), requestedFacts)

	expectedError := entities.FactGatheringError{
		Type: "sbd-config-file-error",
		Message: "error reading sbd configuration file: " +
			"could not open sbd config file: open /path/to/some-non-existent-sbd-config: no such file or directory",
	}

	suite.EqualError(err, expectedError.Error())
	suite.Empty(gatheredFacts)
}

func (suite *SBDGathererTestSuite) TestInvalidConfigFile() {
	gatherer := gatherers.NewSBDGatherer(helpers.GetFixturePath("discovery/cluster/sbd/sbd_config_invalid"))

	gatheredFacts, err := gatherer.Gather(context.Background(), []entities.FactRequest{})

	expectedError := &entities.FactGatheringError{
		Type:    "sbd-config-file-error",
		Message: "error reading sbd configuration file: could not parse sbd config file: error on line 1: missing =",
	}

	suite.EqualError(err, expectedError.Error())
	suite.Empty(gatheredFacts)
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
			Name:     "sbd_integer_value",
			Gatherer: "sbd_config",
			Argument: "AN_INTEGER",
		},
		{
			Name:     "sbd_empty_value",
			Gatherer: "sbd_config",
			Argument: "SBD_OPTS",
		},
		{
			Name:     "sbd_unexistent",
			Gatherer: "sbd_config",
			Argument: "SBD_THIS_DOES_NOT_EXIST",
		},
	}

	gatherer := gatherers.NewSBDGatherer(helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"))

	gatheredFacts, err := gatherer.Gather(context.Background(), requestedFacts)

	expectedFacts := []entities.Fact{
		{
			Name:  "sbd_pacemaker",
			Value: &entities.FactValueString{Value: "yes"},
		},
		{
			Name:  "sbd_startmode",
			Value: &entities.FactValueString{Value: "always"},
		},
		{
			Name:  "sbd_integer_value",
			Value: &entities.FactValueInt{Value: 42},
		},
		{
			Name:  "sbd_empty_value",
			Value: &entities.FactValueString{Value: ""},
		},
		{
			Name: "sbd_unexistent",
			Error: &entities.FactGatheringError{
				Type:    "sbd-config-value-not-found",
				Message: "requested field value not found: SBD_THIS_DOES_NOT_EXIST",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedFacts, gatheredFacts)
}
