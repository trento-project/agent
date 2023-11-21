package collector_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/test/helpers"
)

const (
	DummyAgentID = "779cdd70-e9e2-58ca-b18a-bf3eb3f71244"
)

type CollectorClientTestSuite struct {
	suite.Suite
	collectorClient *collector.Collector
	httpClient      *http.Client
}

func TestCollectorClientTestSuite(t *testing.T) {
	suite.Run(t, new(CollectorClientTestSuite))
}

func (suite *CollectorClientTestSuite) SetupSuite() {
	httpClient := http.DefaultClient
	collectorClient := collector.NewCollectorClient(
		&collector.Config{
			AgentID:   DummyAgentID,
			ServerURL: "https://localhost",
			APIKey:    apiKey,
		},
		httpClient,
	)

	suite.collectorClient = collectorClient
	suite.httpClient = httpClient
}

func (suite *CollectorClientTestSuite) TestCollectorClientPublishingSuccess() {
	ctx := context.TODO()
	discoveredDataPayload := struct {
		FieldA string
	}{
		FieldA: "some discovered field",
	}

	discoveryType := "the_discovery_type"

	suite.httpClient.Transport = helpers.RoundTripFunc(func(req *http.Request) *http.Response {
		requestBody, err := json.Marshal(map[string]interface{}{
			"agent_id":       DummyAgentID,
			"discovery_type": discoveryType,
			"payload":        discoveredDataPayload,
		})

		suite.NoError(err)
		bodyBytes, _ := io.ReadAll(req.Body)

		suite.EqualValues(requestBody, bodyBytes)

		suite.Equal(req.URL.String(), "https://localhost/api/v1/collect")
		return &http.Response{
			StatusCode: 202,
		}
	})

	err := suite.collectorClient.Publish(ctx, discoveryType, discoveredDataPayload)

	suite.NoError(err)
}

func (suite *CollectorClientTestSuite) TestCollectorClientPublishingFailure() {
	ctx := context.TODO()
	suite.httpClient.Transport = helpers.RoundTripFunc(func(req *http.Request) *http.Response {
		suite.Equal(req.URL.String(), "https://localhost/api/v1/collect")
		return &http.Response{
			StatusCode: 500,
		}
	})

	err := suite.collectorClient.Publish(ctx, "some_discovery_type", struct{}{})

	suite.Error(err)
}

func (suite *CollectorClientTestSuite) TestCollectorClientHeartbeat() {
	ctx := context.TODO()
	suite.httpClient.Transport = helpers.RoundTripFunc(func(req *http.Request) *http.Response {
		suite.Equal(req.URL.String(), fmt.Sprintf("https://localhost/api/v1/hosts/%s/heartbeat", DummyAgentID))
		return &http.Response{
			StatusCode: 204,
		}
	})
	err := suite.collectorClient.Heartbeat(ctx)

	suite.NoError(err)
}
