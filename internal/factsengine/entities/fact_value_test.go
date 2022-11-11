package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFactValueAsInterface(t *testing.T) {
	cases := []struct {
		description string
		factValue   FactValue
		expected    interface{}
	}{
		{
			description: "FactValueInt AsInterface",
			factValue:   &FactValueInt{Value: 1},
			expected:    1,
		},
		{
			description: "FactValueFloat AsInterface",
			factValue:   &FactValueFloat{Value: 1.1},
			expected:    1.1,
		},
		{
			description: "FactValueString AsInterface",
			factValue:   &FactValueString{Value: "test"},
			expected:    "test",
		},
		{
			description: "FactValueBool AsInterface",
			factValue:   &FactValueBool{Value: true},
			expected:    true,
		},
		{
			description: "FactValueMap AsInterface",
			factValue:   &FactValueMap{Value: map[string]FactValue{"test": &FactValueString{Value: "test"}}},
			expected:    map[string]interface{}{"test": "test"},
		},
		{
			description: "FactValueList AsInterface",
			factValue:   &FactValueList{Value: []FactValue{&FactValueString{Value: "test"}}},
			expected:    []interface{}{"test"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			i := tt.factValue.AsInterface()

			assert.Equal(t, i, tt.expected)
		})
	}
}

func TestParseStringToFactValue(t *testing.T) {
	cases := []struct {
		description string
		str         string
		expected    FactValue
	}{
		{
			description: "Should parse a string to FactValueInt",
			str:         "1",
			expected:    &FactValueInt{Value: 1},
		},
		{
			description: "Should parse a string to  FactValueFloat",
			str:         "1.1",
			expected:    &FactValueFloat{Value: 1.1},
		},

		{
			description: "Should parse a string to FactValueBool",
			str:         "true",
			expected:    &FactValueBool{Value: true},
		},
		{
			description: "Should parse a string to FactValueString",
			str:         "test",
			expected:    &FactValueString{Value: "test"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			factValue := ParseStringToFactValue(tt.str)

			assert.Equal(t, factValue, tt.expected)
		})
	}
}
