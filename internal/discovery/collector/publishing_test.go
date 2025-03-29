//nolint:lll
package collector_test

import (
	"context"
	"encoding/json"
	"fmt"
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
	apiKey             = "some-api-key"
	sapSystemdiscovery = "sap_system_discovery"
	cloudDiscovery     = "cloud_discovery"
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
	discoveryType := "ha_cluster_discovery"
	discoveredCluster := mocks.NewDiscoveredClusterMock()

	suite.runDiscoveryScenario(discoveryType, discoveredCluster, func(requestBodyAgainstCollector string) {
		suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath("discovery/cluster/expected_published_cluster_discovery.json"), requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClientPublishingCloudDiscovery() {
	discoveryType := "cloud_discovery"
	discoveredCloudInstance := mocks.NewDiscoveredCloudMock()

	suite.runDiscoveryScenario(discoveryType, discoveredCloudInstance, func(requestBodyAgainstCollector string) {
		suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath("discovery/azure/expected_published_cloud_discovery.json"), requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClientPublishingHostDiscovery() {
	discoveryType := "host_discovery"
	discoveredHost := mocks.NewDiscoveredHostMock()

	suite.runDiscoveryScenario(discoveryType, discoveredHost, func(requestBodyAgainstCollector string) {
		suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath("discovery/host/expected_published_host_discovery.json"), requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClientPublishingSubscriptionDiscovery() {
	discoveredSubscriptions := mocks.NewDiscoveredSubscriptionsMock()
	discoveryType := "subscription_discovery"
	suite.runDiscoveryScenario(discoveryType, discoveredSubscriptions, func(requestBodyAgainstCollector string) {
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

func (suite *PublishingTestSuite) TestCollectorClientPublishingAWSCloudDiscoveryWithoutMetadata() {
	discoveredCloudInstance := cloud.Instance{
		Provider: cloud.AWS,
		Metadata: nil,
	}

	suite.runDiscoveryScenario(cloudDiscovery, discoveredCloudInstance, func(requestBodyAgainstCollector string) {
		suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath("discovery/aws/collector_payloads/expected_published_cloud_discovery_without_metadata.json"), requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClientPublishinCloudDiscoveryWithoutMetadata() {
	providersWithoutMetadata := []string{
		cloud.Nutanix,
		cloud.KVM,
		cloud.VMware,
	}

	for _, provider := range providersWithoutMetadata {
		discoveredInstance := cloud.Instance{
			Provider: provider,
			Metadata: nil,
		}

		path := fmt.Sprintf("discovery/provider/%s/expected_published_cloud_discovery.json", provider)

		suite.runDiscoveryScenario(cloudDiscovery, discoveredInstance, func(requestBodyAgainstCollector string) {
			suite.assertJSONMatchesJSONFileContent(helpers.GetFixturePath(path), requestBodyAgainstCollector)
		})
	}
}

type AssertionFunc func(requestBodyAgainstCollector string)

func (suite *PublishingTestSuite) runDiscoveryScenario(discoveryType string, payload interface{}, assertion AssertionFunc) {
	ctx := context.TODO()
	collectorClient := suite.configuredClient

	suite.httpClient.Transport = helpers.RoundTripFunc(func(req *http.Request) *http.Response {
		requestBody, err := json.Marshal(map[string]interface{}{
			"agent_id":       DummyAgentID,
			"discovery_type": discoveryType,
			"payload":        payload,
		})

		suite.NoError(err)

		outgoingRequestBody, _ := io.ReadAll(req.Body)

		suite.EqualValues(requestBody, outgoingRequestBody)

		assertion(string(outgoingRequestBody))

		suite.Equal(req.URL.String(), "https://localhost/api/v1/collect")
		suite.Equal(req.Header.Get("X-Trento-apiKey"), apiKey)
		return &http.Response{ //nolint
			StatusCode: 202,
		}
	})

	err := collectorClient.Publish(ctx, discoveryType, payload)

	suite.NoError(err)
}

func (suite *PublishingTestSuite) assertJSONMatchesJSONFileContent(expectedJSONContentPath string, actualJSON string) {
	expectedJSONContent, err := os.Open(expectedJSONContentPath)
	if err != nil {
		panic(err)
	}

	b, _ := io.ReadAll(expectedJSONContent)

	suite.JSONEq(string(b), actualJSON)
}
