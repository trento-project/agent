package saptune_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/saptune"
	"github.com/trento-project/agent/pkg/utils/mocks"
)

type SaptuneTestSuite struct {
	suite.Suite
}

func TestSaptuneTestSuite(t *testing.T) {
	suite.Run(t, new(SaptuneTestSuite))
}

func (suite *SaptuneTestSuite) TestNewSaptune() {
	mockCommand := new(mocks.MockCommandExecutor)

	mockCommand.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.1.0"), nil,
	)

	saptuneRetriever, err := saptune.NewSaptune(mockCommand)

	suite.NoError(err)
	suite.Equal("3.1.0", saptuneRetriever.Version)
	suite.Equal(true, saptuneRetriever.IsJSONSupported)
}

func (suite *SaptuneTestSuite) TestNewSaptuneUnsupportedSaptuneVer() {
	mockCommand := new(mocks.MockCommandExecutor)

	mockCommand.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.0.0"), nil,
	)

	saptuneRetriever, err := saptune.NewSaptune(mockCommand)

	suite.NoError(err)
	suite.Equal("3.0.0", saptuneRetriever.Version)
	suite.Equal(false, saptuneRetriever.IsJSONSupported)
}

func (suite *SaptuneTestSuite) TestNewSaptuneSaptuneVersionUnknownErr() {
	mockCommand := new(mocks.MockCommandExecutor)

	mockCommand.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		nil, errors.New("Error: exec: \"rpm\": executable file not found in $PATH"),
	)

	saptuneRetriever, err := saptune.NewSaptune(mockCommand)

	suite.EqualError(err, saptune.ErrSaptuneVersionUnknown.Error()+": Error: exec: \"rpm\": executable file not found in $PATH")
	suite.Equal("", saptuneRetriever.Version)
	suite.Equal(false, saptuneRetriever.IsJSONSupported)
}

func (suite *SaptuneTestSuite) TestRunCommand() {
	mockCommand := new(mocks.MockCommandExecutor)

	saptuneOutput := []byte("some_output")

	mockCommand.
		On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").
		Return([]byte("3.0.0"), nil).
		On("Exec", "saptune", "some_command").
		Return(saptuneOutput, nil)

	saptuneRetriever, _ := saptune.NewSaptune(mockCommand)

	statusOutput, err := saptuneRetriever.RunCommand("some_command")

	expectedOutput := []byte("some_output")

	suite.NoError(err)
	suite.Equal(expectedOutput, statusOutput)
}

func (suite *SaptuneTestSuite) TestRunCommandJSON() {
	mockCommand := new(mocks.MockCommandExecutor)

	saptuneOutput := []byte("{\"some_json_key\": \"some_value\"}")

	mockCommand.
		On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").
		Return([]byte("3.1.0"), nil).
		On("Exec", "saptune", "--format", "json", "status").
		Return(saptuneOutput, nil)

	saptuneRetriever, _ := saptune.NewSaptune(mockCommand)

	statusOutput, err := saptuneRetriever.RunCommandJSON("status")

	expectedOutput := []byte("{\"some_json_key\": \"some_value\"}")

	suite.NoError(err)
	suite.Equal(expectedOutput, statusOutput)
}

func (suite *SaptuneTestSuite) TestRunCommandJSONNoJSONSupported() {
	mockCommand := new(mocks.MockCommandExecutor)

	mockCommand.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.0.0"), nil,
	)

	saptuneRetriever, _ := saptune.NewSaptune(mockCommand)

	statusOutput, err := saptuneRetriever.RunCommandJSON("status")

	expectedOutput := []byte(nil)

	suite.EqualError(err, saptune.ErrUnsupportedSaptuneVer.Error())
	suite.Equal(expectedOutput, statusOutput)
}
