package factsengine

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

type GatheringTestSuite struct {
	suite.Suite
}

func TestGatheringTestSuite(t *testing.T) {
	suite.Run(t, new(GatheringTestSuite))
}

type DummyGatherer1 struct {
}

func NewDummyGatherer1() *DummyGatherer1 {
	return &DummyGatherer1{}
}

func (s *DummyGatherer1) Gather(_ []gatherers.FactRequest) ([]gatherers.Fact, error) {
	return []gatherers.Fact{
		{
			Name:    "dummy1",
			Value:   "1",
			CheckID: "check1",
		},
	}, nil
}

type DummyGatherer2 struct {
}

func NewDummyGatherer2() *DummyGatherer2 {
	return &DummyGatherer2{}
}

func (s *DummyGatherer2) Gather(_ []gatherers.FactRequest) ([]gatherers.Fact, error) {
	return []gatherers.Fact{
		{
			Name:    "dummy2",
			Value:   "2",
			CheckID: "check1",
		},
	}, nil
}

type ErrorGatherer struct {
}

func NewErrorGatherer() *ErrorGatherer {
	return &ErrorGatherer{}
}

func (s *ErrorGatherer) Gather(_ []gatherers.FactRequest) ([]gatherers.Fact, error) {
	return []gatherers.Fact{}, fmt.Errorf("kabum!") //nolint
}

func (suite *GatheringTestSuite) TestGatheringGatherFacts() {
	someID := "someID"     //nolint
	agentID := "someAgent" //nolint

	factsRequest := gatherers.FactsRequest{
		ExecutionID: someID,
		Facts: []gatherers.FactRequest{
			{
				Name:     "dummy1",
				Gatherer: "dummyGatherer1",
				Argument: "dummy1",
				CheckID:  "check1",
			},
			{
				Name:     "dummy2",
				Gatherer: "dummyGatherer2",
				Argument: "dummy2",
				CheckID:  "check1",
			},
		},
	}

	factGatherers := map[string]gatherers.FactGatherer{
		"dummyGatherer1": NewDummyGatherer1(),
		"dummyGatherer2": NewDummyGatherer2(),
	}

	factResults, err := gatherFacts(agentID, factsRequest, factGatherers)

	expectedFacts := []gatherers.Fact{
		{
			Name:    "dummy1",
			Value:   "1",
			CheckID: "check1",
		},
		{
			Name:    "dummy2",
			Value:   "2",
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.Equal(someID, factResults.ExecutionID)
	suite.Equal(agentID, factResults.AgentID)
	suite.ElementsMatch(expectedFacts, factResults.Facts)
}

func (suite *GatheringTestSuite) TestFactsEngineGatherFactsGathererNotFound() {
	someID := "someID"
	agentID := "someAgent"

	factsRequest := gatherers.FactsRequest{
		ExecutionID: someID,
		Facts: []gatherers.FactRequest{
			{
				Name:     "dummy1",
				Gatherer: "dummyGatherer1",
				Argument: "dummy1",
				CheckID:  "check1",
			},
			{
				Name:     "other",
				Gatherer: "otherGatherer",
				Argument: "other",
				CheckID:  "check1",
			},
		},
	}

	factGatherers := map[string]gatherers.FactGatherer{
		"dummyGatherer1": NewDummyGatherer1(),
		"dummyGatherer2": NewDummyGatherer2(),
	}

	factResults, err := gatherFacts(agentID, factsRequest, factGatherers)

	expectedFacts := []gatherers.Fact{
		{
			Name:    "dummy1",
			Value:   "1",
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.Equal(someID, factResults.ExecutionID)
	suite.Equal(agentID, factResults.AgentID)
	suite.ElementsMatch(expectedFacts, factResults.Facts)
}

func (suite *GatheringTestSuite) TestFactsEngineGatherFactsErrorGathering() {
	someID := "someID"
	agentID := "someAgent"

	factsRequest := gatherers.FactsRequest{
		ExecutionID: someID,
		Facts: []gatherers.FactRequest{
			{
				Name:     "dummy1",
				Gatherer: "dummyGatherer1",
				Argument: "dummy1",
				CheckID:  "check1",
			},
			{
				Name:     "error",
				Gatherer: "errorGatherer",
				Argument: "error",
				CheckID:  "check1",
			},
		},
	}

	factGatherers := map[string]gatherers.FactGatherer{
		"dummyGatherer1": NewDummyGatherer1(),
		"errorGatherer":  NewErrorGatherer(),
	}

	factResults, err := gatherFacts(agentID, factsRequest, factGatherers)

	expectedFacts := []gatherers.Fact{
		{
			Name:    "dummy1",
			Value:   "1",
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.Equal(someID, factResults.ExecutionID)
	suite.Equal(agentID, factResults.AgentID)
	suite.ElementsMatch(expectedFacts, factResults.Facts)
}
