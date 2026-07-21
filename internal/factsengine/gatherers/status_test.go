// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package gatherers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const testAgentID = "779cdd70-e9e2-58ca-b18a-bf3eb3f71244"

type StatusGathererTestSuite struct {
	suite.Suite
}

func TestStatusGathererTestSuite(t *testing.T) {
	suite.Run(t, new(StatusGathererTestSuite))
}

func (suite *StatusGathererTestSuite) TestStatusGathererSuccess() {
	g := gatherers.NewStatusGatherer(testAgentID)

	factRequests := []entities.FactRequest{
		{
			Name:     "status",
			Gatherer: "status@v1",
			CheckID:  "check1",
		},
	}

	factResults, err := g.Gather(context.Background(), factRequests)

	suite.NoError(err)
	suite.Len(factResults, 1)

	resultMap, ok := factResults[0].Value.(*entities.FactValueMap)
	suite.Require().True(ok)
	suite.Equal("check1", factResults[0].CheckID)
	agentID, ok := resultMap.Value["agent_id"].(*entities.FactValueString)
	suite.Require().True(ok)
	suite.Equal(testAgentID, agentID.Value)
	suite.IsType(&entities.FactValueString{}, resultMap.Value["version"])
}

func (suite *StatusGathererTestSuite) TestStatusGathererMultipleRequests() {
	g := gatherers.NewStatusGatherer(testAgentID)

	factRequests := []entities.FactRequest{
		{Name: "status", Gatherer: "status@v1", CheckID: "check1"},
		{Name: "status", Gatherer: "status@v1", CheckID: "check2"},
	}

	factResults, err := g.Gather(context.Background(), factRequests)

	suite.NoError(err)
	suite.Len(factResults, 2)
	suite.Equal(factResults[0].Value, factResults[1].Value)
}

func (suite *StatusGathererTestSuite) TestStatusGathererContextCancelled() {
	g := gatherers.NewStatusGatherer(testAgentID)

	factsRequest := []entities.FactRequest{{
		Name:     "status",
		Gatherer: "status@v1",
		CheckID:  "check1",
	}}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	factResults, err := g.Gather(ctx, factsRequest)

	suite.Error(err)
	suite.Empty(factResults)
}
