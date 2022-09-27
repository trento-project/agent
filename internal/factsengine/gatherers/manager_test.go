package gatherers_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
)

type TestManager struct {
	suite.Suite
}

func TestTestManager(t *testing.T) {
	suite.Run(t, new(TestManager))
}

func (suite *TestManager) TestManagerGetGatherer() {
	manager := gatherers.NewManager(map[string]gatherers.FactGatherer{
		gatherers.CorosyncFactKey: gatherers.NewDefaultCorosyncConfGatherer(),
	})

	r, err := manager.GetGatherer(gatherers.CorosyncFactKey)
	expectedGatherer := gatherers.NewDefaultCorosyncConfGatherer()

	suite.NoError(err)
	suite.Equal(expectedGatherer, r)
}

func (suite *TestManager) TestManagerAddGatherers() {
	manager := gatherers.NewManager(map[string]gatherers.FactGatherer{
		gatherers.CorosyncFactKey: gatherers.NewDefaultCorosyncConfGatherer(),
	})

	manager.AddGatherers(map[string]gatherers.FactGatherer{
		"test": &mocks.FactGatherer{},
	})

	expectedGatherers := []string{gatherers.CorosyncFactKey, "test"}

	// we sort the array in order to have consistency in the tests
	// map keys are not ordered ofc

	result := manager.AvailableGatherers()
	sort.Strings(result)

	suite.Equal(expectedGatherers, result)
}

func (suite *TestManager) TestFactsEngineGetGathererNotFound() {
	manager := gatherers.NewManager(map[string]gatherers.FactGatherer{
		gatherers.CorosyncFactKey: gatherers.NewDefaultCorosyncConfGatherer(),
	})
	_, err := manager.GetGatherer("other")

	suite.EqualError(err, "gatherer other not found")
}

func (suite *TestManager) TestManagerAvailableGatherers() {
	manager := gatherers.NewManager(map[string]gatherers.FactGatherer{
		"dummyGatherer1": &mocks.FactGatherer{},
		"dummyGatherer2": &mocks.FactGatherer{},
		"errorGatherer":  &mocks.FactGatherer{},
	})

	gatherers := manager.AvailableGatherers()

	expectedGatherers := []string{"dummyGatherer1", "dummyGatherer2", "errorGatherer"}

	suite.ElementsMatch(expectedGatherers, gatherers)
}
