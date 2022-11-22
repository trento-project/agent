package entities_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

type FactValueTestSuite struct {
	suite.Suite
}

func TestFactValueTestSuite(t *testing.T) {
	suite.Run(t, new(FactValueTestSuite))
}

func (suite *FactValueTestSuite) TestFactValueAsInterface() {
	cases := []struct {
		description string
		factValue   entities.FactValue
		expected    interface{}
	}{
		{
			description: "FactValueInt AsInterface",
			factValue:   &entities.FactValueInt{Value: 1},
			expected:    1,
		},
		{
			description: "FactValueFloat AsInterface",
			factValue:   &entities.FactValueFloat{Value: 1.1},
			expected:    1.1,
		},
		{
			description: "FactValueString AsInterface",
			factValue:   &entities.FactValueString{Value: "test"},
			expected:    "test",
		},
		{
			description: "FactValueBool AsInterface",
			factValue:   &entities.FactValueBool{Value: true},
			expected:    true,
		},
		{
			description: "FactValueMap AsInterface",
			factValue: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"test": &entities.FactValueString{Value: "test"}}},
			expected: map[string]interface{}{"test": "test"},
		},
		{
			description: "FactValueList AsInterface",
			factValue: &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueString{Value: "test"}}},
			expected: []interface{}{"test"},
		},
	}

	for _, tt := range cases {
		suite.T().Run(tt.description, func(t *testing.T) {
			i := tt.factValue.AsInterface()

			suite.Equal(i, tt.expected)
		})
	}
}

func (suite *FactValueTestSuite) TestParseStringToFactValue() {
	cases := []struct {
		description string
		str         string
		expected    entities.FactValue
	}{
		{
			description: "Should parse a string to FactValueInt",
			str:         "1",
			expected:    &entities.FactValueInt{Value: 1},
		},
		{
			description: "Should parse a string to  FactValueFloat",
			str:         "1.1",
			expected:    &entities.FactValueFloat{Value: 1.1},
		},

		{
			description: "Should parse a string to FactValueBool",
			str:         "true",
			expected:    &entities.FactValueBool{Value: true},
		},
		{
			description: "Should parse a string to FactValueString",
			str:         "test",
			expected:    &entities.FactValueString{Value: "test"},
		},
	}

	for _, tt := range cases {
		suite.T().Run(tt.description, func(t *testing.T) {
			factValue := entities.ParseStringToFactValue(tt.str)

			suite.Equal(factValue, tt.expected)
		})
	}
}
