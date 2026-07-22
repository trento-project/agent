// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package collector_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cloud"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/internal/discovery/mocks"
	"github.com/trento-project/agent/test/helpers"
)

const (
	apiKey                = "some-api-key"
	hostDiscovery         = "host_discovery"
	sapSystemdiscovery    = "sap_system_discovery"
	cloudDiscovery        = "cloud_discovery"
	clusterDiscovery      = "ha_cluster_discovery"
	subscriptionDiscovery = "subscription_discovery"
)

type PublishingTestSuite struct {
	suite.Suite

	configuredClient *collector.Collector
	httpClient       *http.Client
}

func TestPublishingTestSuite(t *testing.T) {
	suite.Run(t, new(PublishingTestSuite))
}

func (suite *PublishingTestSuite) SetupSuite() {
	httpClient := http.DefaultClient
	collectorClient := collector.NewCollectorClient(
		&collector.Config{
			AgentID:   DummyAgentID,
			ServerURL: "https://localhost",
			APIKey:    apiKey,
		},
		httpClient,
	)

	suite.configuredClient = collectorClient
	suite.httpClient = httpClient
}

// Following test cover publishing data from the discovery loops

func (suite *PublishingTestSuite) TestCollectorClientPublishingClusterDiscovery() {
	discoveredCluster := mocks.NewDiscoveredClusterMock()

	suite.runDiscoveryScenario(clusterDiscovery, discoveredCluster, func(requestBodyAgainstCollector string) {
		suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath("discovery/cluster/expected_published_cluster_discovery.json"), requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClientPublishingCloudDiscovery() {
	discoveredCloudInstance := mocks.NewDiscoveredCloudMock()

	suite.runDiscoveryScenario(cloudDiscovery, discoveredCloudInstance, func(requestBodyAgainstCollector string) {
		suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath("discovery/azure/expected_published_cloud_discovery.json"), requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClientPublishingHostDiscovery() {
	discoveredHost := mocks.NewDiscoveredHostMock()

	suite.runDiscoveryScenario(hostDiscovery, discoveredHost, func(requestBodyAgainstCollector string) {
		suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath("discovery/host/expected_published_host_discovery.json"), requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClientPublishingSubscriptionDiscovery() {
	discoveredSubscriptions := mocks.NewDiscoveredSubscriptionsMock()
	suite.runDiscoveryScenario(subscriptionDiscovery, discoveredSubscriptions, func(requestBodyAgainstCollector string) {
		suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath("discovery/subscriptions/expected_published_subscriptions_discovery.json"), requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClientPublishingSAPSystemDatabaseDiscovery() {
	discoveredSAPSystem := mocks.NewDiscoveredSAPSystemDatabaseMock()

	suite.runDiscoveryScenario(sapSystemdiscovery, discoveredSAPSystem, func(requestBodyAgainstCollector string) {
		suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath("discovery/sap_system/expected_published_sap_system_discovery_database.json"), requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClientPublishingSAPSystemApplicationDiscovery() {
	discoveredSAPSystem := mocks.NewDiscoveredSAPSystemApplicationMock()

	suite.runDiscoveryScenario(sapSystemdiscovery, discoveredSAPSystem, func(requestBodyAgainstCollector string) {
		suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath("discovery/sap_system/expected_published_sap_system_discovery_application.json"), requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClientPublishingSAPSystemDiagnosticsDiscovery() {
	discoveredSAPSystem := mocks.NewDiscoveredSAPSystemDiagnosticsMock()

	suite.runDiscoveryScenario(sapSystemdiscovery, discoveredSAPSystem, func(requestBodyAgainstCollector string) {
		suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath("discovery/sap_system/expected_published_sap_system_discovery_diagnostics.json"), requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClientPublishingAWSCloudDiscoveryWithMetadata() {
	discoveredCloudInstance := cloud.Instance{
		Provider: cloud.AWS,
		Metadata: &cloud.AWSMetadata{
			AmiID:              "ami-123456",
			BlockDeviceMapping: map[string]string{"device": "volume"},
			InstanceID:         "i-123456",
			InstanceType:       "t2.micro",
			Placement: cloud.Placement{
				AvailabilityZone: "us-west-2a",
				Region:           "us-west-2",
			},
		},
	}

	suite.runDiscoveryScenario(cloudDiscovery, discoveredCloudInstance, func(requestBodyAgainstCollector string) {
		suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath("discovery/aws/collector_payloads/expected_published_cloud_discovery_with_metadata.json"), requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClientPublishingCloudDiscoveryWithoutMetadata() {
	providersWithoutMetadata := map[string]string{
		cloud.AWS:     "discovery/aws/collector_payloads/expected_published_cloud_discovery_without_metadata.json",
		cloud.Nutanix: "discovery/provider/nutanix/expected_published_cloud_discovery.json",
		cloud.KVM:     "discovery/provider/kvm/expected_published_cloud_discovery.json",
		cloud.VMware:  "discovery/provider/vmware/expected_published_cloud_discovery.json",
	}

	for provider, fixture := range providersWithoutMetadata {
		discoveredInstance := cloud.Instance{
			Provider: provider,
			Metadata: nil,
		}

		suite.runDiscoveryScenario(cloudDiscovery, discoveredInstance, func(requestBodyAgainstCollector string) {
			suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath(fixture), requestBodyAgainstCollector)
		})
	}
}

type AssertionFunc func(requestBodyAgainstCollector string)

func (suite *PublishingTestSuite) runDiscoveryScenario(discoveryType string, payload any, assertion AssertionFunc) {
	ctx := context.TODO()
	collectorClient := suite.configuredClient

	suite.httpClient.Transport = helpers.RoundTripFunc(func(req *http.Request) *http.Response {
		requestBody, err := json.Marshal(map[string]any{
			"agent_id":       DummyAgentID,
			"discovery_type": discoveryType,
			"payload":        payload,
		})

		suite.Require().NoError(err)

		outgoingRequestBody, _ := io.ReadAll(req.Body)

		suite.Equal(requestBody, outgoingRequestBody)

		assertion(string(outgoingRequestBody))

		suite.Equal("https://localhost/api/v1/collect", req.URL.String())
		suite.Equal(apiKey, req.Header.Get("X-Trento-Apikey"))

		return &http.Response{
			StatusCode: http.StatusAccepted,
		}
	})

	err := collectorClient.Publish(ctx, discoveryType, payload)

	suite.Require().NoError(err)
}

func (suite *PublishingTestSuite) assertJSONMatchesJSONFileContent(expectedJSONContentPath string, actualJSON string) {
	expectedJSONContent, err := os.Open(expectedJSONContentPath)
	if err != nil {
		panic(err)
	}

	b, _ := io.ReadAll(expectedJSONContent)

	suite.JSONEq(string(b), actualJSON)
}
