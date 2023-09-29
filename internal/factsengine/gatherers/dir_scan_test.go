package gatherers_test

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const dirScanTestBasePath = "/var/test"
const dirScanTestGlobPattern = "/var/test/*/ASCS*"

type DirScanGathererSuite struct {
	suite.Suite
	testFS afero.Fs
}

func TestDirScanGathererSuite(t *testing.T) {
	suite.Run(t, new(DirScanGathererSuite))
}

func (s *DirScanGathererSuite) SetupSuite() {
	tFs := afero.NewMemMapFs()

	for i := 0; i <= 2; i++ {
		path := fmt.Sprintf("%s/%d/ASCS%d", dirScanTestBasePath, i, i)
		_, _ = tFs.Create(path)
	}

	s.testFS = tFs
}

func (s *DirScanGathererSuite) TestDirScanningSuccess() {
	g := gatherers.NewDirScanGatherer(s.testFS)

	fr := []entities.FactRequest{{
		Argument: dirScanTestGlobPattern,
		CheckID:  "check1",
		Gatherer: "dir-scan",
		Name:     "dir-scan",
	}}

	expectedResult := []entities.Fact{{
		Name:    "dir-scan",
		CheckID: "check1",
		Value: &entities.FactValueMap{
			Value: map[string]entities.FactValue{
				"/var/test/0": &entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"owner": &entities.FactValueInt{Value: 0},
						"group": &entities.FactValueInt{Value: 0},
						"files": &entities.FactValueList{
							Value: []entities.FactValue{
								&entities.FactValueString{Value: "/var/test/0/ASCS0"},
							},
						},
					},
				},
				"/var/test/1": &entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"owner": &entities.FactValueInt{Value: 0},
						"group": &entities.FactValueInt{Value: 0},
						"files": &entities.FactValueList{
							Value: []entities.FactValue{
								&entities.FactValueString{Value: "/var/test/1/ASCS1"},
							},
						},
					},
				},
				"/var/test/2": &entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"owner": &entities.FactValueInt{Value: 0},
						"group": &entities.FactValueInt{Value: 0},
						"files": &entities.FactValueList{
							Value: []entities.FactValue{
								&entities.FactValueString{Value: "/var/test/2/ASCS2"},
							},
						},
					},
				},
			},
		},
	}}

	result, err := g.Gather(fr)
	s.NoError(err)
	s.EqualValues(expectedResult, result)
}
