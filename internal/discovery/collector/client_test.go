// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

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
		requestBody, err := json.Marshal(map[string]any{
			"agent_id":       DummyAgentID,
			"discovery_type": discoveryType,
			"payload":        discoveredDataPayload,
		})

		suite.Require().NoError(err)

		bodyBytes, _ := io.ReadAll(req.Body)

		suite.Equal(requestBody, bodyBytes)

		suite.Equal("https://localhost/api/v1/collect", req.URL.String())

		return &http.Response{
			StatusCode: http.StatusAccepted,
		}
	})

	err := suite.collectorClient.Publish(ctx, discoveryType, discoveredDataPayload)

	suite.Require().NoError(err)
}

func (suite *CollectorClientTestSuite) TestCollectorClientPublishingFailure() {
	ctx := context.TODO()
	suite.httpClient.Transport = helpers.RoundTripFunc(func(req *http.Request) *http.Response {
		suite.Equal("https://localhost/api/v1/collect", req.URL.String())

		return &http.Response{
			StatusCode: http.StatusInternalServerError,
		}
	})

	err := suite.collectorClient.Publish(ctx, "some_discovery_type", struct{}{})

	suite.Require().Error(err)
}

func (suite *CollectorClientTestSuite) TestCollectorClientHeartbeat() {
	ctx := context.TODO()
	suite.httpClient.Transport = helpers.RoundTripFunc(func(req *http.Request) *http.Response {
		suite.Equal(req.URL.String(), fmt.Sprintf("https://localhost/api/v1/hosts/%s/heartbeat", DummyAgentID))

		return &http.Response{
			StatusCode: http.StatusNoContent,
		}
	})
	err := suite.collectorClient.Heartbeat(ctx)

	suite.Require().NoError(err)
}
