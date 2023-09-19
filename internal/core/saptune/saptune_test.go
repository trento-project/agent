package saptune

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/pkg/utils/mocks"
)

type SaptuneTestSuite struct {
	suite.Suite
}

func TestSaptuneTestSuite(t *testing.T) {
	suite.Run(t, new(SaptuneTestSuite))
}

func (suite *SaptuneTestSuite) TestNewSaptune() {
	mockCommand := new(mocks.CommandExecutor)

	mockCommand.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.1.0"), nil,
	)

	saptuneRetriever, err := NewSaptune(mockCommand)

	expectedSaptune := Saptune{
		Version:         "3.1.0",
		IsJSONSupported: true,
		executor:        mockCommand,
	}

	suite.NoError(err)
	suite.Equal(expectedSaptune, saptuneRetriever)
}

func (suite *SaptuneTestSuite) TestNewSaptuneUnsupportedSaptuneVer() {
	mockCommand := new(mocks.CommandExecutor)

	mockCommand.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.0.0"), nil,
	)

	saptuneRetriever, err := NewSaptune(mockCommand)

	expectedSaptune := Saptune{
		Version:         "3.0.0",
		IsJSONSupported: false,
		executor:        mockCommand,
	}

	suite.NoError(err)
	suite.Equal(expectedSaptune, saptuneRetriever)
}

func (suite *SaptuneTestSuite) TestNewSaptuneSaptuneVersionUnknownErr() {
	mockCommand := new(mocks.CommandExecutor)

	mockCommand.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		nil, errors.New("Error: exec: \"rpm\": executable file not found in $PATH"),
	)

	saptuneRetriever, err := NewSaptune(mockCommand)

	expectedSaptune := Saptune{
		Version:         "",
		IsJSONSupported: false,
	}

	suite.EqualError(err, ErrSaptuneVersionUnknown.Error()+": Error: exec: \"rpm\": executable file not found in $PATH")
	suite.Equal(expectedSaptune, saptuneRetriever)
}

func (suite *SaptuneTestSuite) TestRunCommand() {
	mockCommand := new(mocks.CommandExecutor)

	saptuneRetriever := Saptune{
		Version:         "3.0.0",
		IsJSONSupported: false,
		executor:        mockCommand,
	}

	saptuneOutput := []byte("some_output")

	mockCommand.On("Exec", "saptune", "some_command").Return(
		saptuneOutput, nil,
	)

	statusOutput, err := saptuneRetriever.RunCommand("some_command")

	expectedOutput := []byte("some_output")

	suite.NoError(err)
	suite.Equal(expectedOutput, statusOutput)
}

func (suite *SaptuneTestSuite) TestRunCommandJSON() {
	mockCommand := new(mocks.CommandExecutor)

	saptuneRetriever := Saptune{
		Version:         "3.1.0",
		IsJSONSupported: true,
		executor:        mockCommand,
	}

	saptuneOutput := []byte("{\"some_json_key\": \"some_value\"}")

	mockCommand.On("Exec", "saptune", "--format", "json", "status").Return(
		saptuneOutput, nil,
	)

	statusOutput, err := saptuneRetriever.RunCommandJSON("status")

	expectedOutput := []byte("{\"some_json_key\": \"some_value\"}")

	suite.NoError(err)
	suite.Equal(expectedOutput, statusOutput)
}

func (suite *SaptuneTestSuite) TestRunCommandJSONNoJSONSupported() {
	mockCommand := new(mocks.CommandExecutor)

	saptuneRetriever := Saptune{
		IsJSONSupported: false,
		executor:        mockCommand,
	}

	statusOutput, err := saptuneRetriever.RunCommandJSON("status")

	expectedOutput := []byte(nil)

	suite.EqualError(err, ErrUnsupportedSaptuneVer.Error())
	suite.Equal(expectedOutput, statusOutput)
}
