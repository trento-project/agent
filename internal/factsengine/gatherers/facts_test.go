package gatherers

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

	factsGathered := FactsResult{
		ExecutionID: someID,
		AgentID:     someAgent,
		Facts: []Fact{
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
func (suite *FactsTestSuite) TestFactsPrettifyFactResult() {
	fact := Fact{
		Name:    "some-fact",
		Value:   1,
		CheckID: "check1",
	}

	prettifiedFact, err := PrettifyFactResult(fact)

	expectedResponse := "{\n  \"Name\": \"some-fact\",\n  \"Value\": 1,\n  \"CheckID\": \"check1\"\n}"

	suite.NoError(err)
	suite.Equal(expectedResponse, prettifiedFact)
}
