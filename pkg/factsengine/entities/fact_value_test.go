package entities_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type FactValueTestSuite struct {
	suite.Suite
}

func TestFactValueTestSuite(t *testing.T) {
	suite.Run(t, new(FactValueTestSuite))
}

func (suite *FactValueTestSuite) TestNewFactValueWithStringConversion() {
	cases := []struct {
		description string
		factValue   interface{}
		expected    entities.FactValue
		err         error
	}{
		{
			description: "Should construct a basic type to FactValue",
			factValue:   "1",
			expected:    &entities.FactValueInt{Value: 1},
			err:         nil,
		},
		{
			description: "Should construct a list type to FactValue",
			factValue:   []interface{}{"string", 2},
			expected: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueString{Value: "string"},
				&entities.FactValueInt{Value: 2},
			}},
			err: nil,
		},
		{
			description: "Should construct a map type to FactValue",
			factValue: map[string]interface{}{
				"basic": "basic",
				"list":  []interface{}{"string", 2, []interface{}{1.5}},
				"map": map[string]interface{}{
					"int": 5,
				},
			},
			expected: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"basic": &entities.FactValueString{Value: "basic"},
					"list": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueString{Value: "string"},
							&entities.FactValueInt{Value: 2},
							&entities.FactValueList{Value: []entities.FactValue{
								&entities.FactValueFloat{Value: 1.5},
							}},
						}},
					"map": &entities.FactValueMap{Value: map[string]entities.FactValue{
						"int": &entities.FactValueInt{Value: 5},
					}},
				}},
			err: nil,
		},
		{
			description: "Should fail on basic unknown type",
			factValue:   nil,
			expected:    nil,
			err:         fmt.Errorf("invalid type: %T for value: %v", nil, nil),
		},
		{
			description: "Should fail if a list contains an unknown type",
			factValue:   []interface{}{"string", nil},
			expected:    nil,
			err:         fmt.Errorf("invalid type: %T for value: %v", nil, nil),
		},
		{
			description: "Should fail if a map contains an unknown type",
			factValue: map[string]interface{}{
				"basic": &entities.FactValueString{Value: "basic"},
				"nil":   nil,
			},
			expected: nil,
			err:      fmt.Errorf("invalid type: %T for value: %v", nil, nil),
		},
	}

	for _, tt := range cases {
		suite.T().Run(tt.description, func(t *testing.T) {
			factValue, err := entities.NewFactValue(tt.factValue, entities.WithStringConversion())

			suite.Equal(tt.expected, factValue)
			suite.Equal(tt.err, err)
		})
	}
}

func (suite *FactValueTestSuite) TestNewFactValueWithSnakeCaseKeys() {
	inputFactValue := map[string]interface{}{
		"some Key":    "value",
		"MyCamelCase": []interface{}{"string", "2", []interface{}{"1.5"}},
		"map": map[string]interface{}{
			"int": 5,
		},
	}

	factValue, err := entities.NewFactValue(
		inputFactValue,
		entities.WithSnakeCaseKeys())

	expected := &entities.FactValueMap{
		Value: map[string]entities.FactValue{
			"some_key": &entities.FactValueString{Value: "value"},
			"my_camel_case": &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueString{Value: "string"},
					&entities.FactValueString{Value: "2"},
					&entities.FactValueList{Value: []entities.FactValue{
						&entities.FactValueString{Value: "1.5"},
					}},
				}},
			"map": &entities.FactValueMap{Value: map[string]entities.FactValue{
				"int": &entities.FactValueInt{Value: 5},
			}},
		}}

	suite.Equal(expected, factValue)
	suite.NoError(err)
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

func (suite *FactValueTestSuite) TestFactValueListAppend() {
	list := entities.FactValueList{Value: []entities.FactValue{
		&entities.FactValueInt{Value: 1},
	}}
	list.AppendValue(&entities.FactValueInt{Value: 2})

	expected := entities.FactValueList{Value: []entities.FactValue{
		&entities.FactValueInt{Value: 1},
		&entities.FactValueInt{Value: 2},
	}}

	suite.Equal(list, expected)
}

func (suite *FactValueTestSuite) TestFactValueMapGetValue() {
	mapValue := &entities.FactValueMap{
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
			"empty_entry": &entities.FactValueString{},
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
			expected:    mapValue,
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
			description: "Should return ValueNotFoundError when value does not exist because the entry is empty",
			key:         "empty_entry.some_value",
			expected:    nil,
			err: &entities.FactGatheringError{
				Type:    "value-not-found",
				Message: "error getting value: requested field value not found: empty_entry.some_value",
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
			factValue, err := mapValue.GetValue(tt.key)

			suite.Equal(factValue, tt.expected)
			suite.Equal(err, tt.err)
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
			description: "Should parse a string to FactValueFloat",
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
		{
			description: "Should parse float infinity values to FactValueString",
			str:         "INFINITY",
			expected:    &entities.FactValueString{Value: "INFINITY"},
		},
	}

	for _, tt := range cases {
		suite.T().Run(tt.description, func(t *testing.T) {
			factValue := entities.ParseStringToFactValue(tt.str)

			suite.Equal(factValue, tt.expected)
		})
	}
}
