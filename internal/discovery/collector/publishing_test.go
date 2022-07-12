package collector

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/discovery/mocks"
	_ "github.com/trento-project/agent/test"
	"github.com/trento-project/agent/test/helpers"
)

type PublishingTestSuite struct {
	suite.Suite
	configuredClient *client
}

func TestPublishingTestSuite(t *testing.T) {
	suite.Run(t, new(PublishingTestSuite))
}

func (suite *PublishingTestSuite) SetupSuite() {
	collectorClient := NewCollectorClient(
		&Config{
			AgentID:   DummyAgentID,
			ServerUrl: "https://localhost",
			ApiKey:    "some-api-key",
		})

	suite.configuredClient = collectorClient
}

// Following test cover publishing data from the discovery loops

func (suite *PublishingTestSuite) TestCollectorClient_PublishingClusterDiscovery() {
	discoveryType := "ha_cluster_discovery"
	discoveredCluster := mocks.NewDiscoveredClusterMock()

	suite.runDiscoveryScenario(discoveryType, discoveredCluster, func(requestBodyAgainstCollector string) {
		suite.assertJsonMatchesJsonFileContent("./test/fixtures/discovery/cluster/expected_published_cluster_discovery.json", requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClient_PublishingCloudDiscovery() {
	discoveryType := "cloud_discovery"
	discoveredCloudInstance := mocks.NewDiscoveredCloudMock()

	suite.runDiscoveryScenario(discoveryType, discoveredCloudInstance, func(requestBodyAgainstCollector string) {
		suite.assertJsonMatchesJsonFileContent("./test/fixtures/discovery/azure/expected_published_cloud_discovery.json", requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClient_PublishingHostDiscovery() {
	discoveryType := "host_discovery"
	discoveredHost := mocks.NewDiscoveredHostMock()

	suite.runDiscoveryScenario(discoveryType, discoveredHost, func(requestBodyAgainstCollector string) {
		suite.assertJsonMatchesJsonFileContent("./test/fixtures/discovery/host/expected_published_host_discovery.json", requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClient_PublishingSubscriptionDiscovery() {
	discoveryType := "subscription_discovery"
	discoveredSubscriptions := mocks.NewDiscoveredSubscriptionsMock()

	suite.runDiscoveryScenario(discoveryType, discoveredSubscriptions, func(requestBodyAgainstCollector string) {
		suite.assertJsonMatchesJsonFileContent("./test/fixtures/discovery/subscriptions/expected_published_subscriptions_discovery.json", requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClient_PublishingSAPSystemDatabaseDiscovery() {
	discoveryType := "sap_system_discovery"
	discoveredSAPSystem := mocks.NewDiscoveredSAPSystemDatabaseMock()

	suite.runDiscoveryScenario(discoveryType, discoveredSAPSystem, func(requestBodyAgainstCollector string) {
		suite.assertJsonMatchesJsonFileContent("./test/fixtures/discovery/sap_system/expected_published_sap_system_discovery_database.json", requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClient_PublishingSAPSystemApplicationDiscovery() {
	discoveryType := "sap_system_discovery"
	discoveredSAPSystem := mocks.NewDiscoveredSAPSystemApplicationMock()

	suite.runDiscoveryScenario(discoveryType, discoveredSAPSystem, func(requestBodyAgainstCollector string) {
		suite.assertJsonMatchesJsonFileContent("./test/fixtures/discovery/sap_system/expected_published_sap_system_discovery_application.json", requestBodyAgainstCollector)
	})
}

func (suite *PublishingTestSuite) TestCollectorClient_PublishingSAPSystemDiagnosticsDiscovery() {
	discoveryType := "sap_system_discovery"
	discoveredSAPSystem := mocks.NewDiscoveredSAPSystemDiagnosticsMock()

	suite.runDiscoveryScenario(discoveryType, discoveredSAPSystem, func(requestBodyAgainstCollector string) {
		suite.assertJsonMatchesJsonFileContent("./test/fixtures/discovery/sap_system/expected_published_sap_system_discovery_diagnostics.json", requestBodyAgainstCollector)
	})
}

type AssertionFunc func(requestBodyAgainstCollector string)

func (suite *PublishingTestSuite) runDiscoveryScenario(discoveryType string, payload interface{}, assertion AssertionFunc) {
	collectorClient := suite.configuredClient

	collectorClient.httpClient.Transport = helpers.RoundTripFunc(func(req *http.Request) *http.Response {
		requestBody, _ := json.Marshal(map[string]interface{}{
			"agent_id":       DummyAgentID,
			"discovery_type": discoveryType,
			"payload":        payload,
		})

		outgoingRequestBody, _ := ioutil.ReadAll(req.Body)

		suite.EqualValues(requestBody, outgoingRequestBody)

		assertion(string(outgoingRequestBody))

		suite.Equal(req.URL.String(), "https://localhost/api/collect")
		suite.Equal(req.Header.Get("X-Trento-apiKey"), suite.configuredClient.config.ApiKey)
		return &http.Response{
			StatusCode: 202,
		}
	})

	err := collectorClient.Publish(discoveryType, payload)

	suite.NoError(err)
}

func (suite *PublishingTestSuite) assertJsonMatchesJsonFileContent(expectedJsonContentPath string, actualJson string) {
	expectedJsonContent, err := os.Open(expectedJsonContentPath)
	if err != nil {
		panic(err)
	}

	b, _ := ioutil.ReadAll(expectedJsonContent)

	suite.JSONEq(string(b), actualJson)
}
