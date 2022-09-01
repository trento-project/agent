package factsengine

import (
	"encoding/json"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/adapters/mocks"
	"github.com/trento-project/agent/internal/factsengine/gatherers"

	contracts "github.com/trento-project/contracts/go/pkg/gen/entities"
)

type PolicyTestSuite struct {
	suite.Suite
}

func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyTestSuite))
}

func (suite *PolicyTestSuite) TestPolicyPublishFacts() {
	mockAdatper := new(mocks.Adapter)

	someID := uuid.New().String()
	someAgent := uuid.New().String()

	gatheredFacts := gatherers.FactsResult{
		ExecutionID: someID,
		AgentID:     someAgent,
		Facts: []gatherers.Fact{
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

	mockAdatper.On(
		"Publish",
		cloudevents.ApplicationCloudEventsJSON,
		mock.MatchedBy(func(body []byte) bool {
			var event cloudevents.Event
			if err := json.Unmarshal(body, &event); err != nil {
				panic(err)
			}

			var expectedGatheredFacts gatherers.FactsResult
			err := json.Unmarshal(event.DataEncoded, &expectedGatheredFacts)
			if err != nil {
				panic(err)
			}

			f := contracts.FactsGatheredV1{} // nolint

			suite.Equal(f.Type(), event.Context.GetType())
			suite.Equal(f.Source(), event.Context.GetSource())
			suite.Equal(cloudevents.ApplicationJSON, event.Context.GetDataContentType())
			suite.Equal(expectedGatheredFacts, gatheredFacts)

			return true
		})).Return(nil)

	f := FactsEngine{ // nolint
		agentID:             someAgent,
		factsServiceAdapter: mockAdatper,
		factGatherers:       map[string]gatherers.FactGatherer{},
	}

	err := f.publishFacts(gatheredFacts)

	suite.NoError(err)
}
