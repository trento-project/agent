package operator_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/operations/operator"
	"github.com/trento-project/agent/internal/operations/operator/mocks"
)

type RegistryTest struct {
	suite.Suite
}

func TestRegistryTest(t *testing.T) {
	suite.Run(t, new(RegistryTest))
}

func (suite *RegistryTest) TestRegistryAvailableOperators() {
	registry := operator.NewRegistry(operator.BuildersTree{
		operator.SaptuneApplySolutionOperatorName: map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator { return nil },
			"v2": func(_ string, _ operator.Arguments) operator.Operator { return nil },
		},
		"test": map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator { return nil },
			"v2": func(_ string, _ operator.Arguments) operator.Operator { return nil },
		},
		"test2": map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator { return nil },
		},
	})

	expectedOperators := []string{
		"saptuneapplysolution - v1/v2",
		"test - v1/v2",
		"test2 - v1",
	}

	// we sort the array in order to have consistency in the tests
	// map keys are not ordered ofc

	result := registry.AvailableOperators()
	sort.Strings(result)

	suite.Equal(expectedOperators, result)
}

func (suite *RegistryTest) TestGetOperatorBuilderNotFound() {
	registry := operator.NewRegistry(operator.BuildersTree{
		operator.SaptuneApplySolutionOperatorName: map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator { return nil },
			"v2": func(_ string, _ operator.Arguments) operator.Operator { return nil },
		},
		"test": map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator { return nil },
			"v2": func(_ string, _ operator.Arguments) operator.Operator { return nil },
		},
		"test2": map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator { return nil },
		},
	})
	_, err := registry.GetOperatorBuilder("other@v1")

	suite.EqualError(err, "operator other@v1 not found")
}

func (suite *RegistryTest) TestGetOperatorBuilderFoundWithVersion() {
	foundOperator := mocks.NewMockOperator(suite.T())
	registry := operator.NewRegistry(operator.BuildersTree{
		operator.SaptuneApplySolutionOperatorName: map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator { return foundOperator },
			"v2": func(_ string, _ operator.Arguments) operator.Operator { return nil },
		},
		"test": map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator { return nil },
			"v2": func(_ string, _ operator.Arguments) operator.Operator { return nil },
		},
		"test2": map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator { return nil },
		},
	})

	b, err := registry.GetOperatorBuilder("saptuneapplysolution@v1")

	suite.NoError(err)
	suite.Equal(b("", nil), foundOperator)
}

func (suite *RegistryTest) TestGetOperatorBuilderFoundWithoutVersionGetLast() {
	foundOperator := mocks.NewMockOperator(suite.T())
	registry := operator.NewRegistry(operator.BuildersTree{
		operator.SaptuneApplySolutionOperatorName: map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator { return nil },
			"v2": func(_ string, _ operator.Arguments) operator.Operator { return foundOperator },
		},
		"test": map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator { return nil },
			"v2": func(_ string, _ operator.Arguments) operator.Operator { return nil },
		},
		"test2": map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator { return nil },
		},
	})

	b, err := registry.GetOperatorBuilder("saptuneapplysolution")

	suite.NoError(err)
	suite.Equal(b("", nil), foundOperator)
}
