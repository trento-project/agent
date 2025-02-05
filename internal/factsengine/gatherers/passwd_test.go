package gatherers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/test/helpers"
)

type PasswdTestSuite struct {
	suite.Suite
}

func TestPasswdTestSuite(t *testing.T) {
	suite.Run(t, new(PasswdTestSuite))
}

func (suite *PasswdTestSuite) TestPasswd() {
	c := gatherers.NewPasswdGatherer(helpers.GetFixturePath("gatherers/passwd.basic"))

	factRequests := []entities.FactRequest{
		{
			Name:     "passwd",
			Gatherer: "passwd",
			CheckID:  "check1",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "passwd",
			Value: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"user":        &entities.FactValueString{Value: "bin"},
					"uid":         &entities.FactValueInt{Value: 1},
					"gid":         &entities.FactValueInt{Value: 1},
					"description": &entities.FactValueString{Value: "bin"},
					"home":        &entities.FactValueString{Value: "/bin"},
					"shell":       &entities.FactValueString{Value: "/sbin/nologin"},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"user":        &entities.FactValueString{Value: "chrony"},
					"uid":         &entities.FactValueInt{Value: 474},
					"gid":         &entities.FactValueInt{Value: 475},
					"description": &entities.FactValueString{Value: "Chrony Daemon"},
					"home":        &entities.FactValueString{Value: "/var/lib/chrony"},
					"shell":       &entities.FactValueString{Value: "/bin/false"},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"user":        &entities.FactValueString{Value: "pgadmin"},
					"uid":         &entities.FactValueInt{Value: 441},
					"gid":         &entities.FactValueInt{Value: 440},
					"description": &entities.FactValueString{Value: "pgadmin4"},
					"home":        &entities.FactValueString{Value: "/var/lib/pgadmin"},
					"shell":       &entities.FactValueString{Value: "/bin/false"},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"user":        &entities.FactValueString{Value: "uuidd"},
					"uid":         &entities.FactValueInt{Value: 440},
					"gid":         &entities.FactValueInt{Value: 439},
					"description": &entities.FactValueString{Value: "User for uuidd"},
					"home":        &entities.FactValueString{Value: "/var/run/uuidd"},
					"shell":       &entities.FactValueString{Value: "/sbin/nologin"},
				}},
			}},
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PasswdTestSuite) TestPasswdFileNotExists() {
	c := gatherers.NewPasswdGatherer("non_existing_file")

	factRequests := []entities.FactRequest{
		{
			Name:     "passwd",
			Gatherer: "passwd",
		},
	}

	_, err := c.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: passwd-file-error - error reading /etc/passwd file: "+
		"open non_existing_file: no such file or directory")
}

func (suite *PasswdTestSuite) TestPasswdErrorDecoding() {

	c := gatherers.NewPasswdGatherer(helpers.GetFixturePath("gatherers/passwd.invalid"))

	factRequests := []entities.FactRequest{
		{
			Name:     "passwd",
			Gatherer: "passwd",
		},
	}

	_, err := c.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: passwd-file-error - error reading /etc/passwd file: "+
		"invalid passwd file: line 1 entry does not have 7 values")
}

func (suite *PasswdTestSuite) TestPasswdContextCancelled() {
	gatherer := gatherers.NewPasswdGatherer(helpers.GetFixturePath("gatherers/passwd.basic"))

	factsRequest := []entities.FactRequest{{
		Name:     "passwd",
		Gatherer: "passwd",
		CheckID:  "check1",
	}}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	factResults, err := gatherer.Gather(ctx, factsRequest)

	suite.Error(err)
	suite.Empty(factResults)
}
