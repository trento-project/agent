package entities

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	contracts "github.com/trento-project/contracts/go/pkg/gen/entities"
)

type FactsTestSuite struct {
	suite.Suite
}

func TestFactsTestSuite(t *testing.T) {
	suite.Run(t, new(FactsTestSuite))
}

func (suite *FactsTestSuite) TestFactsGatheredToEvent() {
	someID := uuid.New().String()
	someAgent := uuid.New().String()

	factsGathered := FactsGathered{
		ExecutionID: someID,
		AgentID:     someAgent,
		FactsGathered: []FactsGatheredItem{
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
		},
	}

	result := FactsGatheredToEvent(factsGathered)

	expectedFacts := contracts.FactsGatheredV1{
		AgentId:     someAgent,
		ExecutionId: someID,
		FactsGathered: []*contracts.FactsGatheredItems{
			{
				Name:    "dummy1",
				Value:   "1",
				CheckId: "check1",
				Error:   nil,
			},
			{
				Name:    "dummy2",
				Value:   "2",
				CheckId: "check1",
				Error:   nil,
			},
		},
	}

	suite.Equal(expectedFacts, result)
}

func (suite *FactsTestSuite) TestFactsGatheredWithErrorToEvent() {
	someID := uuid.New().String()
	someAgent := uuid.New().String()

	factsGathered := FactsGathered{
		ExecutionID: someID,
		AgentID:     someAgent,
		FactsGathered: []FactsGatheredItem{
			{
				Name:    "dummy1",
				Value:   nil,
				CheckID: "check1",
				Error: &FactGatheringError{
					Message: "some message",
					Type:    "some_type",
				},
			},
			{
				Name:    "dummy2",
				Value:   "2",
				CheckID: "check1",
			},
		},
	}

	result := FactsGatheredToEvent(factsGathered)

	expectedFacts := contracts.FactsGatheredV1{
		AgentId:     someAgent,
		ExecutionId: someID,
		FactsGathered: []*contracts.FactsGatheredItems{
			{
				Name:    "dummy1",
				Value:   nil,
				CheckId: "check1",
				Error: &contracts.Error{
					Message: "some message",
					Type:    "some_type",
				},
			},
			{
				Name:    "dummy2",
				Value:   "2",
				CheckId: "check1",
				Error:   nil,
			},
		},
	}

	suite.Equal(expectedFacts, result)
}

func (suite *FactsTestSuite) TestFactsPrettifyFactsGatheredItem() {
	fact := FactsGatheredItem{
		Name:    "some-fact",
		Value:   1,
		CheckID: "check1",
		Error:   nil,
	}

	prettifiedFact, err := PrettifyFactsGatheredItem(fact)

	expectedResponse := `{
  "Name": "some-fact",
  "CheckID": "check1",
  "Value": 1,
  "Error": null
}`

	suite.NoError(err)
	suite.Equal(expectedResponse, prettifiedFact)
}

func (suite *FactsTestSuite) TestFactsPrettifyFactsGatheredItemWithError() {
	fact := FactsGatheredItem{
		Name:    "some-fact",
		Value:   nil,
		CheckID: "check1",
		Error: &FactGatheringError{
			Message: "some message",
			Type:    "some_type",
		},
	}

	prettifiedFact, err := PrettifyFactsGatheredItem(fact)

	expectedResponse := `{
  "Name": "some-fact",
  "CheckID": "check1",
  "Value": null,
  "Error": {
    "Message": "some message",
    "Type": "some_type"
  }
}`

	suite.NoError(err)
	suite.Equal(expectedResponse, prettifiedFact)
}
