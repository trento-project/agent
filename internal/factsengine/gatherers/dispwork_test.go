package gatherers_test

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type DispWorkGathererTestSuite struct {
	suite.Suite
	fs           afero.Fs
	mockExecutor *utilsMocks.CommandExecutor
}

func TestDispWorkGathererSuite(t *testing.T) {
	suite.Run(t, new(DispWorkGathererTestSuite))
}

func (suite *DispWorkGathererTestSuite) SetupTest() {
	fs := afero.NewMemMapFs()
	err := fs.MkdirAll("/usr/sap/PRD", 0644)
	suite.NoError(err)
	err = fs.MkdirAll("/usr/sap/QAS", 0644)
	suite.NoError(err)
	err = fs.MkdirAll("/usr/sap/QA2", 0644)
	suite.NoError(err)
	err = fs.MkdirAll("/usr/sap/DEV", 0644)
	suite.NoError(err)

	suite.fs = fs
	suite.mockExecutor = new(utilsMocks.CommandExecutor)
}

func (suite *DispWorkGathererTestSuite) TestDispWorkGatheringSuccess() {
	validOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/dispwork-valid.output"))
	validOutput, _ := io.ReadAll(validOutputFile)
	partialOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/dispwork-partial.output"))
	partialOutput, _ := io.ReadAll(partialOutputFile)
	unsortedOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/dispwork-unsorted.output"))
	unsortedOutput, _ := io.ReadAll(unsortedOutputFile)
	suite.mockExecutor.
		On("Exec", "su", "-", "prdadm", "-c", "\"disp+work\"").
		Return(validOutput, nil).
		On("Exec", "su", "-", "qasadm", "-c", "\"disp+work\"").
		Return(partialOutput, nil).
		On("Exec", "su", "-", "qa2adm", "-c", "\"disp+work\"").
		Return(unsortedOutput, nil).
		On("Exec", "su", "-", "devadm", "-c", "\"disp+work\"").
		Return(nil, errors.New("some error"))

	g := gatherers.NewDispWorkGatherer(suite.fs, suite.mockExecutor)

	fr := []entities.FactRequest{
		{
			Name:     "dispwork",
			CheckID:  "check1",
			Gatherer: "disp+work",
		},
	}

	expectedResults := []entities.Fact{{
		Name:    "dispwork",
		CheckID: "check1",
		Value: &entities.FactValueMap{
			Value: map[string]entities.FactValue{
				"PRD": &entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"compilation_mode": &entities.FactValueString{Value: "UNICODE"},
						"kernel_release":   &entities.FactValueString{Value: "753"},
						"patch_number":     &entities.FactValueString{Value: "900"},
					},
				},
				"QAS": &entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"compilation_mode": &entities.FactValueString{Value: ""},
						"kernel_release":   &entities.FactValueString{Value: "753"},
						"patch_number":     &entities.FactValueString{Value: ""},
					},
				},
				"QA2": &entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"compilation_mode": &entities.FactValueString{Value: "UNICODE"},
						"kernel_release":   &entities.FactValueString{Value: "753"},
						"patch_number":     &entities.FactValueString{Value: "900"},
					},
				},
				"DEV": &entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"compilation_mode": &entities.FactValueString{Value: ""},
						"kernel_release":   &entities.FactValueString{Value: ""},
						"patch_number":     &entities.FactValueString{Value: ""},
					},
				},
			},
		},
	}}

	result, err := g.Gather(context.Background(), fr)
	suite.NoError(err)
	suite.EqualValues(expectedResults, result)
}

func (suite *DispWorkGathererTestSuite) TestDispWorkGatheringEmptyFileSystem() {
	g := gatherers.NewDispWorkGatherer(afero.NewMemMapFs(), suite.mockExecutor)

	fr := []entities.FactRequest{
		{
			Name:     "dispwork",
			CheckID:  "check1",
			Gatherer: "disp+work",
		},
	}

	expectedResults := []entities.Fact{{
		Name:    "dispwork",
		CheckID: "check1",
		Value: &entities.FactValueMap{
			Value: map[string]entities.FactValue{},
		},
	}}

	result, err := g.Gather(context.Background(), fr)
	suite.NoError(err)
	suite.EqualValues(expectedResults, result)
}
