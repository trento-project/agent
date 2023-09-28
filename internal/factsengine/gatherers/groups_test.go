package gatherers_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/test/helpers"
)

type GroupsGathererSuite struct {
	suite.Suite
}

func TestGroupsGatherer(t *testing.T) {
	suite.Run(t, new(GroupsGathererSuite))
}

func (s *GroupsGathererSuite) TestGroupsParsingSuccess() {
	gatherer := gatherers.NewGroupsGatherer(helpers.GetFixturePath("gatherers/groups.valid"))

	fr := []entities.FactRequest{{
		Name:     "groups",
		Gatherer: "groups",
		CheckID:  "checkone",
	}}

	expectedFacts := []entities.Fact{
		{
			Name:    "groups",
			CheckID: "checkone",
			Value: &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"name": &entities.FactValueString{
								Value: "root",
							},
							"gid": &entities.FactValueInt{
								Value: 0,
							},
							"users": &entities.FactValueList{
								Value: []entities.FactValue{},
							},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"name": &entities.FactValueString{
								Value: "daemon",
							},
							"gid": &entities.FactValueInt{
								Value: 1,
							},
							"users": &entities.FactValueList{
								Value: []entities.FactValue{},
							},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"name": &entities.FactValueString{
								Value: "adm",
							},
							"gid": &entities.FactValueInt{
								Value: 4,
							},
							"users": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueString{Value: "syslog"},
									&entities.FactValueString{Value: "trento"},
								},
							},
						},
					},
				},
			},
		},
	}
	results, err := gatherer.Gather(fr)
	s.NoError(err)
	s.EqualValues(expectedFacts, results)
}

func (s *GroupsGathererSuite) TestGroupsParsingDecodeErrorInvalidGID() {
	gatherer := gatherers.NewGroupsGatherer(helpers.GetFixturePath("gatherers/groups.invalidgid"))

	fr := []entities.FactRequest{{
		Name:     "groups",
		Gatherer: "groups",
		CheckID:  "checkone",
	}}

	result, err := gatherer.Gather(fr)
	s.Nil(result)
	s.EqualError(err, "fact gathering error: groups-decoding-error - error deconding groups file: could not convert group id  to integer")
}

func (s *GroupsGathererSuite) TestGroupsParsingDecodeErrorInvalidFormat() {
	gatherer := gatherers.NewGroupsGatherer(helpers.GetFixturePath("gatherers/groups.invalidformat"))

	fr := []entities.FactRequest{{
		Name:     "groups",
		Gatherer: "groups",
		CheckID:  "checkone",
	}}

	result, err := gatherer.Gather(fr)
	s.Nil(result)
	s.EqualError(err, "fact gathering error: groups-decoding-error - error deconding groups file: could not decode groups file line daemon:x:1, entry are less then 4")
}
