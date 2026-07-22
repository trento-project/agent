// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package collector_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

func (suite *CollectorClientTestSuite) TestCollectorClientPublishingFailureResponseBodyHandling() {
	readErr := errors.New("boom")
	largeBody := strings.Repeat("a", 5*1024)

	tests := []struct {
		name              string
		body              io.ReadCloser
		expectContains    string
		expectNotContains string
		expectErrorIs     error
	}{
		{
			name:           "response body is included in the error message, quoted",
			body:           io.NopCloser(strings.NewReader(`{"error": "invalid payload"}`)),
			expectContains: strconv.Quote(`{"error": "invalid payload"}`),
		},
		{
			name:              "large response bodies are truncated",
			body:              io.NopCloser(strings.NewReader(largeBody)),
			expectNotContains: largeBody,
		},
		{
			name:          "errors reading the response body are propagated",
			body:          io.NopCloser(&erroringReader{err: readErr}),
			expectErrorIs: readErr,
		},
	}

	for _, tt := range tests {
		ctx := context.TODO()
		suite.httpClient.Transport = helpers.RoundTripFunc(func(req *http.Request) *http.Response {
			suite.Equal(req.URL.String(), "https://localhost/api/v1/collect")
			return &http.Response{
				StatusCode: 500,
				Body:       tt.body,
			}
		})

		err := suite.collectorClient.Publish(ctx, "some_discovery_type", struct{}{})

		suite.Error(err, tt.name)
		if tt.expectContains != "" {
			suite.Contains(err.Error(), tt.expectContains, tt.name)
		}
		if tt.expectNotContains != "" {
			suite.NotContains(err.Error(), tt.expectNotContains, tt.name)
			suite.Less(len(err.Error()), len(tt.expectNotContains), tt.name)
		}
		if tt.expectErrorIs != nil {
			suite.ErrorIs(err, tt.expectErrorIs, tt.name)
		}
	}
}

type erroringReader struct {
	err error
}

func (r *erroringReader) Read([]byte) (int, error) {
	return 0, r.err
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
