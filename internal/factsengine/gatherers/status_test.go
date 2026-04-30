package gatherers_test

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type StatusGathererTestSuite struct {
	suite.Suite
}

func TestStatusGathererTestSuite(t *testing.T) {
	suite.Run(t, new(StatusGathererTestSuite))
}

func (suite *StatusGathererTestSuite) newFsWithMachineID(machineID string) afero.Fs {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "/etc/machine-id", []byte(machineID+"\n"), 0644)
	return fs
}

func (suite *StatusGathererTestSuite) TestStatusGathererSuccess() {
	fs := suite.newFsWithMachineID("abc123")
	g := gatherers.NewStatusGatherer(fs, "/etc/machine-id")

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
	suite.NotEmpty(agentID.Value)
	suite.IsType(&entities.FactValueString{}, resultMap.Value["version"])
}

func (suite *StatusGathererTestSuite) TestStatusGathererMultipleRequests() {
	fs := suite.newFsWithMachineID("abc123")
	g := gatherers.NewStatusGatherer(fs, "/etc/machine-id")

	factRequests := []entities.FactRequest{
		{Name: "status", Gatherer: "status@v1", CheckID: "check1"},
		{Name: "status", Gatherer: "status@v1", CheckID: "check2"},
	}

	factResults, err := g.Gather(context.Background(), factRequests)

	suite.NoError(err)
	suite.Len(factResults, 2)
	suite.Equal(factResults[0].Value, factResults[1].Value)
}

func (suite *StatusGathererTestSuite) TestStatusGathererMachineIDNotFound() {
	fs := afero.NewMemMapFs()
	g := gatherers.NewStatusGatherer(fs, "/etc/machine-id")

	factRequests := []entities.FactRequest{
		{Name: "status", Gatherer: "status@v1"},
	}

	_, err := g.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: status-machine-id-error - error reading machine ID: "+
		"open /etc/machine-id: file does not exist")
}

func (suite *StatusGathererTestSuite) TestStatusGathererContextCancelled() {
	fs := suite.newFsWithMachineID("abc123")
	g := gatherers.NewStatusGatherer(fs, "/etc/machine-id")

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
