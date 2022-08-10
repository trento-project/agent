package factsengine

import (
	"encoding/json"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/adapters/mocks"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

type PolicyTestSuite struct {
	suite.Suite
}

func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyTestSuite))
}

func (suite *PolicyTestSuite) TestPolicyHandleEvent() {
	mockAdatper := new(mocks.Adapter)

	someID := uuid.New().String()
	someAgent := uuid.New().String()

	expectedFactsResponse := gatherers.FactsResult{
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
		mock.MatchedBy(func(body []byte) bool {
			var event cloudevents.Event
			if err := json.Unmarshal(body, &event); err != nil {
				panic(err)
			}

			var factsResult gatherers.FactsResult
			err := json.Unmarshal(event.DataEncoded, &factsResult)
			if err != nil {
				panic(err)
			}

			suite.Equal(factsGatheredEvent, event.Context.GetType())
			suite.Equal(eventSource, event.Context.GetSource())
			suite.Equal(cloudevents.ApplicationJSON, event.Context.GetDataContentType())
			suite.Equal(someID, factsResult.ExecutionID)
			suite.ElementsMatch(expectedFactsResponse.Facts, factsResult.Facts)

			return true
		}),
		cloudevents.ApplicationCloudEventsJSON).Return(nil)

	f := FactsEngine{ // nolint
		agentID:             someAgent,
		factsServiceAdapter: mockAdatper,
		factGatherers: map[string]gatherers.FactGatherer{
			"dummyGatherer1": NewDummyGatherer1(),
			"dummyGatherer2": NewDummyGatherer2(),
		},
	}

	requestEvent := cloudevents.NewEvent()
	requestEvent.SetID(uuid.New().String())
	requestEvent.SetSource(eventSource)
	requestEvent.SetTime(time.Now())
	requestEvent.SetType(factsGatheringRequestEvent)

	factsRequests := &gatherers.FactsRequest{
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

	err := requestEvent.SetData(cloudevents.ApplicationJSON, factsRequests)
	if err != nil {
		panic(err)
	}

	requestEventJSON, err := json.Marshal(requestEvent)
	if err != nil {
		panic(err)
	}

	err = f.handleEvent(cloudevents.ApplicationCloudEventsJSON, requestEventJSON)

	suite.NoError(err)
}

func (suite *PolicyTestSuite) TestPolicyHandleEventInvalidContentType() {
	f := FactsEngine{} // nolint
	err := f.handleEvent(cloudevents.ApplicationJSON, []byte(""))
	suite.EqualError(err, "Error handling event: invalid content type: application/json")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventInvalidEventType() {
	requestEvent := cloudevents.NewEvent()
	requestEvent.SetID(uuid.New().String())
	requestEvent.SetSource(eventSource)
	requestEvent.SetTime(time.Now())
	requestEvent.SetType("other")

	err := requestEvent.SetData(cloudevents.ApplicationJSON, nil)
	if err != nil {
		panic(err)
	}

	requestEventJSON, err := json.Marshal(requestEvent)
	if err != nil {
		panic(err)
	}

	f := FactsEngine{} // nolint
	err = f.handleEvent(cloudevents.ApplicationCloudEventsJSON, requestEventJSON)
	suite.EqualError(err, "Invalid event type: other")
}

func (suite *PolicyTestSuite) TestPolicyBuildResponse() {
	facts := gatherers.FactsResult{
		ExecutionID: "some-id",
		AgentID:     "some-agent",
		Facts: []gatherers.Fact{
			{
				Name:    "fact1",
				Value:   "1",
				CheckID: "check1",
			},
			{
				Name:    "fact2",
				Value:   "2",
				CheckID: "check2",
			},
		},
	}

	response, err := buildFactsGatheredEvent(facts)

	suite.NoError(err)

	var event cloudevents.Event
	if err := json.Unmarshal(response, &event); err != nil {
		panic(err)
	}

	expectedBody := `{"execution_id":"some-id","agent_id":"some-agent","facts":[` +
		`{"name":"fact1","value":"1","check_id":"check1"},{"name":"fact2","value":"2","check_id":"check2"}]}`

	suite.Equal(factsGatheredEvent, event.Context.GetType())
	suite.Equal(eventSource, event.Context.GetSource())
	suite.Equal(cloudevents.ApplicationJSON, event.Context.GetDataContentType())
	suite.Equal(expectedBody, string(event.DataEncoded))
}
