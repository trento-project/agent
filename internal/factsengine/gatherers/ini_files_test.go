package gatherers_test

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type IniFilesTestSuite struct {
	suite.Suite
}

func TestIniFilesTestSuite(t *testing.T) {
	suite.Run(t, new(IniFilesTestSuite))
}

func (suite *IniFilesTestSuite) TestIniFilesGathererEmptyArgumentProvided() {
	c := gatherers.NewDefaultIniFilesGatherer()

	factRequests := []entities.FactRequest{
		{
			Name:     "any",
			Gatherer: "ini_files",
		},
	}

	_, err := c.Gather(context.Background(), factRequests)

	suite.Error(err)
}

func (suite *IniFilesTestSuite) TestIniFilesGathererUnsupportedArgumentProvided() {
	c := gatherers.NewDefaultIniFilesGatherer()

	factRequests := []entities.FactRequest{
		{
			Name:     "any",
			Gatherer: "ini_files",
			Argument: "invalid.ini",
		},
	}

	_, err := c.Gather(context.Background(), factRequests)

	suite.Error(err)
}

func (suite *IniFilesTestSuite) TestIniFilesGathererNoSAPSystemFound() {
	c := gatherers.NewDefaultIniFilesGatherer()

	factRequests := []entities.FactRequest{
		{
			Name:     "global conf",
			Gatherer: "ini_files",
			Argument: "global.ini",
		},
	}

	_, err := c.Gather(context.Background(), factRequests)

	suite.Error(err)
}

func (suite *IniFilesTestSuite) TestIniFilesGathererEmptyGlobalIni() {

	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/usr/sap/S01/SYS/global/hdb/custom/config/global.ini", []byte(""), 0400)
	suite.NoErrorf(err, "error creating content01")

	c := gatherers.NewIniFilesGatherer(fs)

	factRequests := []entities.FactRequest{
		{
			Name:     "global conf",
			Gatherer: "ini_files",
			Argument: "global.ini",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	suite.NoError(err)
	suite.Len(factResults, 1)
	fact, ok := factResults[0].Value.(*entities.FactValueList)
	if !ok {
		suite.Fail("fact value is not a list")
	}
	suite.Nil(factResults[0].Error)
	suite.Equal(
		fact.Value[0],
		&entities.FactValueMap{
			Value: map[string]entities.FactValue{
				"sid": &entities.FactValueString{Value: "S01"},
				"content": &entities.FactValueMap{
					Value: map[string]entities.FactValue{},
				},
			},
		})

}

func (suite *IniFilesTestSuite) TestIniFilesGathererInvalidGlobalIni() {

	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/usr/sap/S01/SYS/global/hdb/custom/config/global.ini", []byte("invalid"), 0400)
	suite.NoErrorf(err, "error creating content01")

	c := gatherers.NewIniFilesGatherer(fs)

	factRequests := []entities.FactRequest{
		{
			Name:     "global conf",
			Gatherer: "ini_files",
			Argument: "global.ini",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	suite.NoError(err)
	suite.Len(factResults, 1)
	suite.NotNil(factResults[0].Error)
}

func (suite *IniFilesTestSuite) TestIniFilesGathererGlobalIniParse() {
	content := `
	key1=value1
	#comment
	[section1]
	key2=value2

	[section2]
	key3=value3
	`

	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/usr/sap/S01/SYS/global/hdb/custom/config/global.ini", []byte(content), 0400)
	suite.NoErrorf(err, "error creating content01")

	c := gatherers.NewIniFilesGatherer(fs)

	factRequests := []entities.FactRequest{
		{
			Name:     "global conf",
			Gatherer: "ini_files",
			Argument: "global.ini",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	suite.NoError(err)
	suite.Len(factResults, 1)
	suite.Empty(factResults[0].Error)
	fact, ok := factResults[0].Value.(*entities.FactValueList)
	if !ok {
		suite.Fail("fact value is not a list")
	}
	suite.Len(fact.Value, 1)
	suite.Equal(
		fact.Value[0],
		&entities.FactValueMap{
			Value: map[string]entities.FactValue{
				"sid": &entities.FactValueString{Value: "S01"},
				"content": &entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"key1": &entities.FactValueString{Value: "value1"},
						"section1": &entities.FactValueMap{
							Value: map[string]entities.FactValue{
								"key2": &entities.FactValueString{Value: "value2"},
							},
						},
						"section2": &entities.FactValueMap{
							Value: map[string]entities.FactValue{
								"key3": &entities.FactValueString{Value: "value3"},
							},
						},
					},
				},
			},
		})

}

func (suite *IniFilesTestSuite) TestIniFilesGathererGlobalIniMultiParse() {
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/usr/sap/S01/SYS/global/hdb/custom/config/global.ini", []byte("key1=value1"), 0400)
	suite.NoErrorf(err, "error creating content01")
	err = afero.WriteFile(fs, "/usr/sap/S02/SYS/global/hdb/custom/config/global.ini", []byte("key2=value2"), 0400)
	suite.NoErrorf(err, "error creating content02")

	c := gatherers.NewIniFilesGatherer(fs)

	factRequests := []entities.FactRequest{
		{
			Name:     "global conf",
			Gatherer: "ini_files",
			Argument: "global.ini",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	suite.NoError(err)
	suite.Len(factResults, 1)
	suite.Empty(factResults[0].Error)
	suite.IsType(&entities.FactValueList{}, factResults[0].Value)
	fact, ok := factResults[0].Value.(*entities.FactValueList)
	if !ok {
		suite.Fail("fact value is not a list")
	}
	suite.Len(fact.Value, 2)
	suite.Equal(
		fact.Value[0],
		&entities.FactValueMap{
			Value: map[string]entities.FactValue{
				"sid": &entities.FactValueString{Value: "S01"},
				"content": &entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"key1": &entities.FactValueString{Value: "value1"},
					},
				},
			},
		})
	suite.Equal(
		fact.Value[1],
		&entities.FactValueMap{
			Value: map[string]entities.FactValue{
				"sid": &entities.FactValueString{Value: "S02"},
				"content": &entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"key2": &entities.FactValueString{Value: "value2"},
					},
				},
			},
		})

}

func (suite *IniFilesTestSuite) TestIniFilesGathererGlobalIniPartialError() {
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/usr/sap/S01/SYS/global/hdb/custom/config/global.ini", []byte("key1=value1"), 0400)
	suite.NoErrorf(err, "error creating content01")
	err = afero.WriteFile(fs, "/usr/sap/S02/SYS/global/hdb/custom/config/global.ini", []byte("invalid"), 0400)
	suite.NoErrorf(err, "error creating content02")

	c := gatherers.NewIniFilesGatherer(fs)

	factRequests := []entities.FactRequest{
		{
			Name:     "global conf",
			Gatherer: "ini_files",
			Argument: "global.ini",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	suite.NoError(err)
	suite.Len(factResults, 1)
	suite.NotNil(factResults[0].Error)
}

func (suite *IniFilesTestSuite) TestIniFilesGathererContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/usr/sap/S01/SYS/global/hdb/custom/config/global.ini", []byte("key1=value1"), 0400)
	suite.NoErrorf(err, "error creating content01")

	c := gatherers.NewIniFilesGatherer(fs)

	factRequests := []entities.FactRequest{
		{
			Name:     "global conf",
			Gatherer: "ini_files",
			Argument: "global.ini",
		},
	}
	factResults, err := c.Gather(ctx, factRequests)

	suite.Error(err)
	suite.Empty(factResults)
}
