package gatherers_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
)

type RegistryTest struct {
	suite.Suite
}

func TestRegistryTest(t *testing.T) {
	suite.Run(t, new(RegistryTest))
}

func (suite *RegistryTest) RegistryTestGetGatherer() {
	registry := gatherers.NewRegistry(map[string]gatherers.FactGatherer{
		gatherers.CorosyncFactKey: gatherers.NewDefaultCorosyncConfGatherer(),
	})

	r, err := registry.GetGatherer(gatherers.CorosyncFactKey)
	expectedGatherer := gatherers.NewDefaultCorosyncConfGatherer()

	suite.NoError(err)
	suite.Equal(expectedGatherer, r)
}

func (suite *RegistryTest) RegistryTestAddGatherers() {
	registry := gatherers.NewRegistry(map[string]gatherers.FactGatherer{
		gatherers.CorosyncFactKey: gatherers.NewDefaultCorosyncConfGatherer(),
	})

	registry.AddGatherers(map[string]gatherers.FactGatherer{
		"test": &mocks.FactGatherer{},
	})

	expectedGatherers := []string{gatherers.CorosyncFactKey, "test"}

	// we sort the array in order to have consistency in the tests
	// map keys are not ordered ofc

	result := registry.AvailableGatherers()
	sort.Strings(result)

	suite.Equal(expectedGatherers, result)
}

func (suite *RegistryTest) TestFactsEngineGetGathererNotFound() {
	registry := gatherers.NewRegistry(map[string]gatherers.FactGatherer{
		gatherers.CorosyncFactKey: gatherers.NewDefaultCorosyncConfGatherer(),
	})
	_, err := registry.GetGatherer("other")

	suite.EqualError(err, "gatherer other not found")
}

func (suite *RegistryTest) RegistryTestAvailableGatherers() {
	registry := gatherers.NewRegistry(map[string]gatherers.FactGatherer{
		"rdummyGatherer1": &mocks.FactGatherer{},
		"dummyGatherer2":  &mocks.FactGatherer{},
		"errorGatherer":   &mocks.FactGatherer{},
	})

	gatherers := registry.AvailableGatherers()

	expectedGatherers := []string{"dummyGatherer1", "dummyGatherer2", "errorGatherer"}

	suite.ElementsMatch(expectedGatherers, gatherers)
}
