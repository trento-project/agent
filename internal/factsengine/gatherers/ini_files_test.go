package gatherers_test

import (
	"context"
	"testing"

	st "github.com/balanza/supertouch"
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

	fs, err := st.Touch(
		st.Tree(
			st.Dir("/usr/sap/S01/SYS/global/hdb/custom/config",
				st.EmptyFile("global.ini"),
			),
		),
		st.WithFileSystem(afero.NewMemMapFs()),
	)
	suite.NoErrorf(err, "error creating fs")

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
	suite.Error(factResults[0].Error)
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

	fs, err := st.Touch(
		st.Tree(
			st.Dir("/usr/sap/S01/SYS/global/hdb/custom/config",
				st.File("global.ini", content),
			),
		),
		st.WithFileSystem(afero.NewMemMapFs()),
	)
	suite.NoErrorf(err, "error creating fs")

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
		})
}

func (suite *IniFilesTestSuite) TestIniFilesGathererGlobalIniMultiParse() {
	content01 := "key1=value1"
	content02 := "key2=value2"

	fs, err := st.Touch(
		st.Tree(
			st.Dir("/usr/sap/S01/SYS/global/hdb/custom/config",
				st.File("global.ini", content01),
			),
			st.Dir("/usr/sap/S02/SYS/global/hdb/custom/config",
				st.File("global.ini", content02),
			),
		),
		st.WithFileSystem(afero.NewMemMapFs()),
	)
	suite.NoErrorf(err, "error creating fs")

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
				"key1": &entities.FactValueString{Value: "value1"},
			},
		})
	suite.Equal(
		fact.Value[1],
		&entities.FactValueMap{
			Value: map[string]entities.FactValue{
				"key2": &entities.FactValueString{Value: "value2"},
			},
		})

}

func (suite *IniFilesTestSuite) TestIniFilesGathererGlobalIniPartial() {
	fs, err := st.Touch(
		st.Tree(
			st.Dir("/usr/sap/S01/SYS/global/hdb/custom/config",
				st.File("global.ini", "key1=value1"),
			),
			st.Dir("/usr/sap/S02/SYS/global/hdb/custom/config",
				st.EmptyFile("global.ini"),
			),
		),
		st.WithFileSystem(afero.NewMemMapFs()),
	)
	suite.NoErrorf(err, "error creating fs")

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
	suite.Error(factResults[0].Error)
}

func (suite *IniFilesTestSuite) TestIniFilesGathererContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fs, err := st.Touch(
		st.Tree(
			st.Dir("/usr/sap/S01/SYS/global/hdb/custom/config",
				st.File("global.ini", "key1=value1"),
			),
		),
		st.WithFileSystem(afero.NewMemMapFs()),
	)
	suite.NoErrorf(err, "error creating fs")

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
