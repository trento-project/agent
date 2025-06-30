package gatherers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
)

type SudoersTestSuite struct {
	suite.Suite
	mockExecutor *utilsMocks.MockCommandExecutor
}

func TestSudoersTestSuite(t *testing.T) {
	suite.Run(t, new(SudoersTestSuite))
}

func (suite *SudoersTestSuite) SetupTest() {
	suite.mockExecutor = new(utilsMocks.MockCommandExecutor)
}

func (suite *SudoersTestSuite) TestSudoersGathererSingleUserFound() {
	mockOutput := []byte(`
User foo_user may run the following commands on host:
	(ALL) ALL
    (ALL) NOPASSWD: /usr/sbin/cmd1 --flag, /usr/sbin/cmd2
    (ALL:ALL) ALL
`)
	suite.mockExecutor.
		On("ExecContext", mock.Anything, "/usr/bin/sudo", "-l", "-U", "foo_user").
		Return(mockOutput, nil).
		Once()

	c := gatherers.NewSudoersGatherer(suite.mockExecutor, afero.NewMemMapFs())

	factRequests := []entities.FactRequest{
		{
			Name:     "any",
			Gatherer: "sudoers",
			Argument: "foo_user",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	suite.NoError(err)
	suite.Len(factResults, 1)
	suite.Empty(factResults[0].Error)
	suite.Equal(
		&entities.FactValueList{
			Value: []entities.FactValue{
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"user":         &entities.FactValueString{Value: "foo_user"},
						"run_as_user":  &entities.FactValueString{Value: "ALL"},
						"run_as_group": &entities.FactValueString{Value: ""},
						"no_password":  &entities.FactValueBool{Value: false},
						"command":      &entities.FactValueString{Value: "ALL"},
					}},
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"user":         &entities.FactValueString{Value: "foo_user"},
						"run_as_user":  &entities.FactValueString{Value: "ALL"},
						"run_as_group": &entities.FactValueString{Value: ""},
						"no_password":  &entities.FactValueBool{Value: true},
						"command":      &entities.FactValueString{Value: "/usr/sbin/cmd1 --flag"},
					}},
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"user":         &entities.FactValueString{Value: "foo_user"},
						"run_as_user":  &entities.FactValueString{Value: "ALL"},
						"run_as_group": &entities.FactValueString{Value: ""},
						"no_password":  &entities.FactValueBool{Value: true},
						"command":      &entities.FactValueString{Value: "/usr/sbin/cmd2"},
					}},
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"user":         &entities.FactValueString{Value: "foo_user"},
						"run_as_user":  &entities.FactValueString{Value: "ALL"},
						"run_as_group": &entities.FactValueString{Value: "ALL"},
						"no_password":  &entities.FactValueBool{Value: false},
						"command":      &entities.FactValueString{Value: "ALL"},
					}},
			},
		},
		factResults[0].Value,
	)
}

func (suite *SudoersTestSuite) TestSudoersGathererSingleUserFoundMultiTags() {
	mockOutput := []byte(`
User foo_user may run the following commands on host:
    (ALL) NOPASSWD:EXEC:  /usr/sbin/cmd1
`)
	suite.mockExecutor.
		On("ExecContext", mock.Anything, "/usr/bin/sudo", "-l", "-U", "foo_user").
		Return(mockOutput, nil).
		Once()

	c := gatherers.NewSudoersGatherer(suite.mockExecutor, afero.NewMemMapFs())

	factRequests := []entities.FactRequest{
		{
			Name:     "any",
			Gatherer: "sudoers",
			Argument: "foo_user",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	suite.NoError(err)
	suite.Len(factResults, 1)
	suite.Empty(factResults[0].Error)
	suite.Equal(
		&entities.FactValueList{
			Value: []entities.FactValue{
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"user":         &entities.FactValueString{Value: "foo_user"},
						"run_as_user":  &entities.FactValueString{Value: "ALL"},
						"run_as_group": &entities.FactValueString{Value: ""},
						"no_password":  &entities.FactValueBool{Value: true},
						"command":      &entities.FactValueString{Value: "/usr/sbin/cmd1"},
					}},
			},
		},
		factResults[0].Value,
	)
}

func (suite *SudoersTestSuite) TestSudoersGathererMultipleUsersFoundMultiTags() {
	mockOutput1 := []byte(`
User fooadm may run the following commands on host:
	(ALL) NOPASSWD: /usr/sbin/cmd1 --flagfoo
`)

	suite.mockExecutor.
		On("ExecContext", mock.Anything, "/usr/bin/sudo", "-l", "-U", "fooadm").
		Return(mockOutput1, nil).
		Once()

	mockOutput2 := []byte(`
User baradm may run the following commands on host:
	(ALL) NOPASSWD: /usr/sbin/cmd1 --flagbar, /usr/sbin/cmd2
`)

	suite.mockExecutor.
		On("ExecContext", mock.Anything, "/usr/bin/sudo", "-l", "-U", "baradm").
		Return(mockOutput2, nil).
		Once()

	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/usr/sap/FOO/SYS/global/hdb/custom/config/global.ini", []byte("key1=value1"), 0400)
	suite.NoErrorf(err, "error creating content01")
	err = afero.WriteFile(fs, "/usr/sap/BAR/SYS/global/hdb/custom/config/global.ini", []byte("key1=value1"), 0400)
	suite.NoErrorf(err, "error creating content02")

	c := gatherers.NewSudoersGatherer(suite.mockExecutor, fs)

	factRequests := []entities.FactRequest{
		{
			Name:     "any",
			Gatherer: "sudoers",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	suite.NoError(err)
	suite.Len(factResults, 1)
	suite.Empty(factResults[0].Error)
	suite.Equal(
		&entities.FactValueList{
			Value: []entities.FactValue{
				// expected for user baradm
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"user":         &entities.FactValueString{Value: "baradm"},
						"run_as_user":  &entities.FactValueString{Value: "ALL"},
						"run_as_group": &entities.FactValueString{Value: ""},
						"no_password":  &entities.FactValueBool{Value: true},
						"command":      &entities.FactValueString{Value: "/usr/sbin/cmd1 --flagbar"},
					}},
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"user":         &entities.FactValueString{Value: "baradm"},
						"run_as_user":  &entities.FactValueString{Value: "ALL"},
						"run_as_group": &entities.FactValueString{Value: ""},
						"no_password":  &entities.FactValueBool{Value: true},
						"command":      &entities.FactValueString{Value: "/usr/sbin/cmd2"},
					}},
				// expected for user fooadm
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"user":         &entities.FactValueString{Value: "fooadm"},
						"run_as_user":  &entities.FactValueString{Value: "ALL"},
						"run_as_group": &entities.FactValueString{Value: ""},
						"no_password":  &entities.FactValueBool{Value: true},
						"command":      &entities.FactValueString{Value: "/usr/sbin/cmd1 --flagfoo"},
					}},
			},
		},
		factResults[0].Value,
	)
}

func (suite *SudoersTestSuite) TestSudoersGathererSingleUserNotFound() {
	mockOutput := []byte(`
sudo: unknown user foo_user
`)
	suite.mockExecutor.
		On("ExecContext", mock.Anything, "/usr/bin/sudo", "-l", "-U", "foo_user").
		Return(mockOutput, nil).
		Once()

	c := gatherers.NewSudoersGatherer(suite.mockExecutor, afero.NewMemMapFs())

	factRequests := []entities.FactRequest{
		{
			Name:     "any",
			Gatherer: "sudoers",
			Argument: "foo_user",
		},
	}

	_, err := c.Gather(context.Background(), factRequests)

	suite.NotEmpty(err)
}

func (suite *SudoersTestSuite) TestSudoersGathererMultipleUsersNotFound() {
	mockOutput1 := []byte(`
sudo: unknown user fooadm
`)

	suite.mockExecutor.
		On("ExecContext", mock.Anything, "/usr/bin/sudo", "-l", "-U", "fooadm").
		Return(mockOutput1, nil).
		Once()

	mockOutput2 := []byte(`
User baradm may run the following commands on host:
	(ALL) NOPASSWD: /usr/sbin/cmd1 --flagbar
`)

	suite.mockExecutor.
		On("ExecContext", mock.Anything, "/usr/bin/sudo", "-l", "-U", "baradm").
		Return(mockOutput2, nil).
		Once()

	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/usr/sap/FOO/SYS/global/hdb/custom/config/global.ini", []byte("key1=value1"), 0400)
	suite.NoErrorf(err, "error creating content01")
	err = afero.WriteFile(fs, "/usr/sap/BAR/SYS/global/hdb/custom/config/global.ini", []byte("key1=value1"), 0400)
	suite.NoErrorf(err, "error creating content02")

	c := gatherers.NewSudoersGatherer(suite.mockExecutor, fs)

	factRequests := []entities.FactRequest{
		{
			Name:     "any",
			Gatherer: "sudoers",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	suite.Nil(err)
	suite.Len(factResults, 1)
	suite.Empty(factResults[0].Error)
	suite.Equal(
		&entities.FactValueList{
			Value: []entities.FactValue{
				// expected for user baradm
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"user":         &entities.FactValueString{Value: "baradm"},
						"run_as_user":  &entities.FactValueString{Value: "ALL"},
						"run_as_group": &entities.FactValueString{Value: ""},
						"no_password":  &entities.FactValueBool{Value: true},
						"command":      &entities.FactValueString{Value: "/usr/sbin/cmd1 --flagbar"},
					}},
			},
		},
		factResults[0].Value,
	)
}

func (suite *SudoersTestSuite) TestSudoersGathererOnError() {

	suite.mockExecutor.
		On("ExecContext", mock.Anything, "/usr/bin/sudo", "-l", "-U", "foo_user").
		Return(nil, errors.New("command failure")).
		Once()

	c := gatherers.NewSudoersGatherer(suite.mockExecutor, afero.NewMemMapFs())

	factRequests := []entities.FactRequest{
		{
			Name:     "any",
			Gatherer: "sudoers",
			Argument: "foo_user",
		},
	}

	_, err := c.Gather(context.Background(), factRequests)

	suite.NotEmpty(err)
}

func (suite *SudoersTestSuite) TestSudoersContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockOutput := []byte(`
User foo_user may run the following commands on host:
	(ALL) ALL
`)
	suite.mockExecutor.
		On("ExecContext", mock.Anything, "/usr/bin/sudo", "-l", "-U", "foo_user").
		Return(mockOutput, nil)

	c := gatherers.NewSudoersGatherer(suite.mockExecutor, afero.NewMemMapFs())

	factRequests := []entities.FactRequest{
		{
			Name:     "any",
			Gatherer: "sudoers",
			Argument: "foo_user",
		},
	}
	factResults, err := c.Gather(ctx, factRequests)

	suite.Error(err)
	suite.Empty(factResults)
}
