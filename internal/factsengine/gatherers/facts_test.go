package gatherers

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type FactsTestSuite struct {
	suite.Suite
}

func TestFactsTestSuite(t *testing.T) {
	suite.Run(t, new(FactsTestSuite))
}

func (suite *FactsTestSuite) TestFactsPrettifyFactResult() {
	fact := Fact{
		Name:    "some-fact",
		Value:   1,
		CheckID: "check1",
	}

	prettifiedFact, err := PrettifyFactResult(fact)

	expectedResponse := "{\n  \"name\": \"some-fact\",\n  \"value\": 1,\n  \"check_id\": \"check1\"\n}"

	suite.NoError(err)
	suite.Equal(expectedResponse, prettifiedFact)
}
