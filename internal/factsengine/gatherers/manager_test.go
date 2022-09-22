package gatherers_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
)

type ManagerTestSuite struct {
	suite.Suite
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

func (suite *ManagerTestSuite) TestManagerGetGatherer() {
	manager := gatherers.NewManager(map[string]gatherers.FactGatherer{
		gatherers.CorosyncFactKey: gatherers.NewDefaultCorosyncConfGatherer(),
	})

	r, err := manager.GetGatherer(gatherers.CorosyncFactKey)
	expectedGatherer := gatherers.NewDefaultCorosyncConfGatherer()

	suite.NoError(err)
	suite.Equal(expectedGatherer, r)
}

func (suite *ManagerTestSuite) TestFactsEngineGetGathererNotFound() {
	manager := gatherers.NewManager(map[string]gatherers.FactGatherer{
		gatherers.CorosyncFactKey: gatherers.NewDefaultCorosyncConfGatherer(),
	})
	_, err := manager.GetGatherer("other")

	suite.EqualError(err, "gatherer other not found")
}

func (suite *ManagerTestSuite) TestFactsEngineGetGatherersList() {
	manager := gatherers.NewManager(map[string]gatherers.FactGatherer{
		"dummyGatherer1": &mocks.FactGatherer{},
		"dummyGatherer2": &mocks.FactGatherer{},
		"errorGatherer":  &mocks.FactGatherer{},
	})

	gatherers := manager.AvailableGatherers()

	expectedGatherers := []string{"dummyGatherer1", "dummyGatherer2", "errorGatherer"}

	suite.ElementsMatch(expectedGatherers, gatherers)
}
