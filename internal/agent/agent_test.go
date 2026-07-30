// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package agent_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/v3/internal/agent"
	"github.com/trento-project/agent/v3/internal/discovery"
	"github.com/trento-project/agent/v3/internal/discovery/collector"
	"github.com/trento-project/agent/v3/test/helpers"
)

type AgentTestSuite struct {
	suite.Suite
}

func TestAgentTestSuite(t *testing.T) {
	suite.Run(t, new(AgentTestSuite))
}

func (suite *AgentTestSuite) TestAgentFailsWithInvalidFactsServiceURL() {
	config := &agent.Config{ //nolint:gosec
		AgentID:      helpers.DummyAgentID,
		InstanceName: "test",
		DiscoveriesConfig: &discovery.DiscoveriesConfig{
			DiscoveriesPeriodsConfig: &discovery.DiscoveriesPeriodConfig{},
			CollectorConfig:          &collector.Config{},
		},
		FactsServiceURL: "amqp://trento:trento@localhost:12345/somevhost",
		PluginsFolder:   "/usr/etc/trento/plugins/",
		PrometheusConfig: &discovery.PrometheusConfig{
			Mode:         "pull",
			ExporterName: "node_exporter",
			Target:       "localhost:9100",
		},
	}

	agent, _ := agent.NewAgent(config)
	ctx := context.Background()
	err := agent.Start(ctx)

	suite.Require().ErrorContains(err, "connect: connection refused")
}
