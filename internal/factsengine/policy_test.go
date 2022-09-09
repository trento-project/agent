package factsengine

import (
	"encoding/json"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/adapters/mocks"
	"github.com/trento-project/agent/internal/factsengine/entities"
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

	mockAdatper.On(
		"Publish",
		factsExchange,
		cloudevents.ApplicationCloudEventsJSON,
		mock.MatchedBy(func(body []byte) bool {
			var event cloudevents.Event
			if err := json.Unmarshal(body, &event); err != nil {
				panic(err)
			}

			facts, err := contracts.NewFactsGatheredV1FromJsonCloudEvent(body)
			if err != nil {
				panic(err)
			}

			f := contracts.FactsGatheredV1{} // nolint

			expectedGatheredFacts := &contracts.FactsGatheredV1{
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

			suite.Equal(f.Type(), event.Context.GetType())
			suite.Equal(f.Source(), event.Context.GetSource())
			suite.Equal(cloudevents.ApplicationJSON, event.Context.GetDataContentType())
			suite.Equal(expectedGatheredFacts, facts)

			return true
		})).Return(nil)

	f := FactsEngine{ // nolint
		agentID:             someAgent,
		factsServiceAdapter: mockAdatper,
		factGatherers:       map[string]gatherers.FactGatherer{},
	}

	gatheredFacts := entities.FactsGathered{
		ExecutionID: someID,
		AgentID:     someAgent,
		FactsGathered: []entities.FactsGatheredItem{
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

	err := f.publishFacts(gatheredFacts)

	suite.NoError(err)
}
