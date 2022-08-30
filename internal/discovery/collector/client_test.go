package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	_ "github.com/trento-project/agent/test"
	"github.com/trento-project/agent/test/helpers"
)

const (
	DummyAgentID = "779cdd70-e9e2-58ca-b18a-bf3eb3f71244"
)

type CollectorClientTestSuite struct {
	suite.Suite
}

func TestCollectorClientTestSuite(t *testing.T) {
	suite.Run(t, new(CollectorClientTestSuite))
}

func (suite *CollectorClientTestSuite) TestCollectorClientPublishingSuccess() {
	collectorClient := NewCollectorClient(
		&Config{
			AgentID:   DummyAgentID,
			ServerURL: "https://localhost",
			APIKey:    "some-api-key",
		})

	discoveredDataPayload := struct {
		FieldA string
	}{
		FieldA: "some discovered field",
	}

	discoveryType := "the_discovery_type"

	collectorClient.httpClient.Transport = helpers.RoundTripFunc(func(req *http.Request) *http.Response {
		requestBody, err := json.Marshal(map[string]interface{}{
			"agent_id":       DummyAgentID,
			"discovery_type": discoveryType,
			"payload":        discoveredDataPayload,
		})

		suite.NoError(err)
		bodyBytes, _ := io.ReadAll(req.Body)

		suite.EqualValues(requestBody, bodyBytes)

		suite.Equal(req.URL.String(), "https://localhost/api/collect")
		return &http.Response{ //nolint
			StatusCode: 202,
		}
	})

	err := collectorClient.Publish(discoveryType, discoveredDataPayload)

	suite.NoError(err)
}

func (suite *CollectorClientTestSuite) TestCollectorClientPublishingFailure() {
	collectorClient := NewCollectorClient(
		&Config{
			AgentID:   DummyAgentID,
			ServerURL: "http://localhost",
			APIKey:    "some-api-key",
		})

	collectorClient.httpClient.Transport = helpers.RoundTripFunc(func(req *http.Request) *http.Response {
		suite.Equal(req.URL.String(), "http://localhost/api/collect")
		return &http.Response{ //nolint
			StatusCode: 500,
		}
	})

	err := collectorClient.Publish("some_discovery_type", struct{}{})

	suite.Error(err)
}

func (suite *CollectorClientTestSuite) TestCollectorClientHeartbeat() {
	collectorClient := NewCollectorClient(
		&Config{
			AgentID:   DummyAgentID,
			ServerURL: "https://localhost",
			APIKey:    "some-api-key",
		})

	collectorClient.httpClient.Transport = helpers.RoundTripFunc(func(req *http.Request) *http.Response {
		suite.Equal(req.URL.String(), fmt.Sprintf("https://localhost/api/hosts/%s/heartbeat", DummyAgentID))
		return &http.Response{ //nolint
			StatusCode: 204,
		}
	})
	err := collectorClient.Heartbeat()

	suite.NoError(err)
}
