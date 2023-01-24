package entities_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type FactsGatheredTestSuite struct {
	suite.Suite
}

func TestFactsGatheredTestSuite(t *testing.T) {
	suite.Run(t, new(FactsGatheredTestSuite))
}

func (suite *FactsGatheredTestSuite) TestFactPrettify() {
	factValueMap := &entities.FactValueMap{
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
		}}

	fact := entities.Fact{
		Name:    "fact",
		CheckID: "12345",
		Value:   factValueMap,
		Error:   nil,
	}

	prettyPrintedOutput, _ := fact.Prettify()

	suite.Equal(prettyPrintedOutput, "Name: fact\nCheck ID: 12345\n\nValue:\n\n#{\n  \"basic\": \"basic\",\n  \"list\": [\n    \"string\",\n    2,\n    [\n      1.5\n    ]\n  ],\n  \"map\": #{\n    \"int\": 5\n  }\n}")
}
