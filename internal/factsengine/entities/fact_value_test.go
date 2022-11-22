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

func (suite *FactValueTestSuite) TestGetValue() {
	parsedValue := &entities.FactValueMap{
		Value: map[string]entities.FactValue{
			"string": &entities.FactValueString{Value: "value"},
			"list_value": &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueInt{Value: 1},
					&entities.FactValueInt{Value: 2},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"id": &entities.FactValueString{Value: "id"},
						},
					},
				},
			},
			"map_value": &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"value": &entities.FactValueString{Value: "other_value"},
				},
			},
		},
	}

	cases := []struct {
		description string
		key         string
		expected    entities.FactValue
		err         *entities.FactGatheringError
	}{
		{
			description: "Should return basic value",
			key:         "string",
			expected:    &entities.FactValueString{Value: "value"},
			err:         nil,
		},
		{
			description: "Should return value from list",
			key:         "list_value.0",
			expected:    &entities.FactValueInt{Value: 1},
			err:         nil,
		},
		{
			description: "Should return second value from list",
			key:         "list_value.1",
			expected:    &entities.FactValueInt{Value: 2},
			err:         nil,
		},
		{
			description: "Should return map value from list",
			key:         "list_value.2.id",
			expected:    &entities.FactValueString{Value: "id"},
			err:         nil,
		},
		{
			description: "Should return complete list",
			key:         "list_value",
			expected: &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueInt{Value: 1},
					&entities.FactValueInt{Value: 2},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"id": &entities.FactValueString{Value: "id"},
						},
					},
				},
			},
			err: nil,
		},
		{
			description: "Should return map value from map",
			key:         "map_value.value",
			expected:    &entities.FactValueString{Value: "other_value"},
			err:         nil,
		},
		{
			description: "Should return complete map when no keys are given",
			key:         "",
			expected:    parsedValue,
			err:         nil,
		},
		{
			description: "Should return ValueNotFoundError when value does not exist",
			key:         "map_value.other_value",
			expected:    nil,
			err: &entities.FactGatheringError{
				Type:    "value-not-found",
				Message: "error getting value: requested field value not found: map_value.other_value",
			},
		},
		{
			description: "Should return ValueNotFoundError when given list index does not exist",
			key:         "list_value.3",
			expected:    nil,
			err: &entities.FactGatheringError{
				Type:    "value-not-found",
				Message: "error getting value: 3 index is not available in the list: list_value.3",
			},
		},
		{
			description: "Should return ValueNotFoundError when given list index is not a number",
			key:         "list_value.x",
			expected:    nil,
			err: &entities.FactGatheringError{
				Type:    "value-not-found",
				Message: "error getting value: list index must be of integer value, x provided: list_value.x",
			},
		},
	}

	for _, tt := range cases {
		suite.T().Run(tt.description, func(t *testing.T) {
			factValue, err := entities.GetValue(parsedValue, tt.key)

			suite.Equal(factValue, tt.expected)
			suite.Equal(err, tt.err)
		})
	}
}
