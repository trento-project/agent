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

func (suite *RegistryTest) TestRegistryAddAndGetGatherers() {
	registry := gatherers.NewRegistry(gatherers.FactGatherersTree{
		gatherers.CorosyncConfGathererName: map[string]gatherers.FactGatherer{
			"v1": &mocks.FactGatherer{},
			"v2": &mocks.FactGatherer{},
		},
	})

	registry.AddGatherers(gatherers.FactGatherersTree{
		"test": map[string]gatherers.FactGatherer{
			"v1": &mocks.FactGatherer{},
			"v2": &mocks.FactGatherer{},
		},
		"test2": map[string]gatherers.FactGatherer{
			"v1": &mocks.FactGatherer{},
		},
	})

	expectedGatherers := []string{
		"corosync.conf - v1/v2",
		"test - v1/v2",
		"test2 - v1",
	}

	// we sort the array in order to have consistency in the tests
	// map keys are not ordered ofc

	result := registry.AvailableGatherers()
	sort.Strings(result)

	suite.Equal(expectedGatherers, result)
}

func (suite *RegistryTest) TestRegistryGetGathererInvalidGathererFormat() {
	registry := gatherers.NewRegistry(gatherers.FactGatherersTree{
		gatherers.CorosyncConfGathererName: map[string]gatherers.FactGatherer{
			"v1": &mocks.FactGatherer{},
			"v2": &mocks.FactGatherer{},
		},
	})

	_, err := registry.GetGatherer("other@v2@v2")

	suite.EqualError(err, "could not extract the gatherer version from other@v2@v2, version should follow <gathererName>@<version> syntax")
}

func (suite *RegistryTest) TestRegistryGetGathererNotFoundWithoutVersion() {
	registry := gatherers.NewRegistry(gatherers.FactGatherersTree{
		gatherers.CorosyncConfGathererName: map[string]gatherers.FactGatherer{
			"v1": &mocks.FactGatherer{},
			"v2": &mocks.FactGatherer{},
		},
	})

	_, err := registry.GetGatherer("other")

	suite.EqualError(err, "gatherer other not found")
}

func (suite *RegistryTest) TestRegistryGetGathererNotFoundWithVersion() {
	registry := gatherers.NewRegistry(gatherers.FactGatherersTree{
		gatherers.CorosyncConfGathererName: map[string]gatherers.FactGatherer{
			"v1": &mocks.FactGatherer{},
			"v2": &mocks.FactGatherer{},
		},
	})

	_, err := registry.GetGatherer("other@v1")

	suite.EqualError(err, "gatherer other@v1 not found")
}

func (suite *RegistryTest) TestRegistryGetGathererFoundWithVersion() {
	expectedGatherer := &mocks.FactGatherer{}
	registry := gatherers.NewRegistry(gatherers.FactGatherersTree{
		"other": map[string]gatherers.FactGatherer{
			"v1": expectedGatherer,
			"v2": &mocks.FactGatherer{},
		},
	})

	result, err := registry.GetGatherer("other@v1")

	suite.NoError(err)
	suite.Equal(expectedGatherer, result)
}

func (suite *RegistryTest) TestRegistryGetGathererFoundWithoutVersion() {
	expectedGatherer := gatherers.NewDefaultFstabGatherer()
	registry := gatherers.NewRegistry(gatherers.FactGatherersTree{
		"other": map[string]gatherers.FactGatherer{
			"v1": &mocks.FactGatherer{},
			"v2": expectedGatherer,
		},
	})

	result, err := registry.GetGatherer("other")

	suite.NoError(err)
	suite.Equal(expectedGatherer, result)
}
