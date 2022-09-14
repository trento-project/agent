// nolint:nosnakecase
package entities

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/trento-project/contracts/golang/pkg/events"
)

type FactsGatheredTestSuite struct {
	suite.Suite
}

func TestFactsGatheredSuite(t *testing.T) {
	suite.Run(t, new(FactsGatheredTestSuite))
}

func (suite *FactsGatheredTestSuite) TestFactsGatheredToEvent() {
	someID := uuid.New().String()
	someAgent := uuid.New().String()

	factsGathered := FactsGathered{
		ExecutionID: someID,
		AgentID:     someAgent,
		FactsGathered: []Fact{
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

	result, err := FactsGatheredToEvent(factsGathered)
	suite.NoError(err)

	var facts events.FactsGathered
	err = events.FromEvent(result, &facts)
	suite.NoError(err)

	expectedFacts := events.FactsGathered{
		AgentId:     someAgent,
		ExecutionId: someID,
		FactsGathered: []*events.Fact{
			{
				Name: "dummy1",
				Value: &events.Fact_TextValue{
					TextValue: "1",
				},
				CheckId: "check1",
			},
			{
				Name: "dummy2",
				Value: &events.Fact_TextValue{
					TextValue: "2",
				},
				CheckId: "check1",
			},
		},
	}

	suite.Equal(expectedFacts.AgentId, facts.AgentId)
	suite.Equal(expectedFacts.ExecutionId, facts.ExecutionId)
	suite.Equal(expectedFacts.FactsGathered, facts.FactsGathered)
}

func (suite *FactsGatheredTestSuite) TestFactsGatheredWithErrorToEvent() {
	someID := uuid.New().String()
	someAgent := uuid.New().String()

	factsGathered := FactsGathered{
		ExecutionID: someID,
		AgentID:     someAgent,
		FactsGathered: []Fact{
			{
				Name:    "dummy1",
				Value:   nil,
				CheckID: "check1",
				Error: &Error{
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

	result, err := FactsGatheredToEvent(factsGathered)
	suite.NoError(err)

	var facts events.FactsGathered
	err = events.FromEvent(result, &facts)
	suite.NoError(err)

	expectedFacts := events.FactsGathered{
		AgentId:     someAgent,
		ExecutionId: someID,
		FactsGathered: []*events.Fact{
			{
				Name: "dummy1",
				Value: &events.Fact_ErrorValue{
					ErrorValue: &events.FactError{
						Message: "some message",
						Type:    "some_type",
					},
				},
				CheckId: "check1",
			},
			{
				Name: "dummy2",
				Value: &events.Fact_TextValue{
					TextValue: "2",
				},
				CheckId: "check1",
			},
		},
	}

	suite.Equal(expectedFacts.AgentId, facts.AgentId)
	suite.Equal(expectedFacts.ExecutionId, facts.ExecutionId)
	suite.Equal(expectedFacts.FactsGathered, facts.FactsGathered)
}

func (suite *FactsGatheredTestSuite) TestFactsPrettifyFactsGatheredItem() {
	fact := Fact{
		Name:    "some-fact",
		Value:   1,
		CheckID: "check1",
		Error:   nil,
	}

	prettifiedFact, err := PrettifyEvent(fact)

	expectedResponse := `{
  "Name": "some-fact",
  "CheckID": "check1",
  "Value": 1,
  "Error": null
}`

	suite.NoError(err)
	suite.Equal(expectedResponse, prettifiedFact)
}

func (suite *FactsGatheredTestSuite) TestFactsPrettifyFactsGatheredItemWithError() {
	fact := Fact{
		Name:    "some-fact",
		Value:   nil,
		CheckID: "check1",
		Error: &Error{
			Message: "some message",
			Type:    "some_type",
		},
	}

	prettifiedFact, err := PrettifyEvent(fact)

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
