package gatherers_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/test/helpers"
)

type FstabGathererTestSuite struct {
	suite.Suite
}

func TestFstabGathererSuite(t *testing.T) {
	suite.Run(t, new(FstabGathererTestSuite))
}

func (s *FstabGathererTestSuite) TestFstabGatheringSuccess() {
	g := gatherers.NewFstabGatherer(helpers.GetFixturePath("gatherers/fstab.valid"))

	fr := []entities.FactRequest{
		{
			Name:     "fstab",
			CheckID:  "check1",
			Gatherer: "fstab",
		},
	}

	result, err := g.Gather(fr)
	s.NoError(err)
}
