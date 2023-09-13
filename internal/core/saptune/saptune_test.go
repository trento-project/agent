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

	saptuneOutput := []byte("{\"some_json_key\": \"some_value\"}")

	mockCommand.On("Exec", "saptune", "--format", "json", "status").Return(
		saptuneOutput, nil,
	)

	saptuneRetriever, err := NewSaptune(mockCommand)
	suite.NoError(err)
	statusOutput, err := saptuneRetriever.RunCommand("--format", "json", "status")

	expectedVersion := "3.1.0"
	expectedOutput := []byte("{\"some_json_key\": \"some_value\"}")

	suite.NoError(err)
	suite.Equal(expectedOutput, statusOutput)
	suite.Equal(expectedVersion, saptuneRetriever.Version)
}

func (suite *SaptuneTestSuite) TestNewSaptuneSaptuneVersionUnknownErr() {
	mockCommand := new(mocks.CommandExecutor)

	mockCommand.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		nil, errors.New("Error: exec: \"rpm\": executable file not found in $PATH"),
	)

	saptuneRetriever, err := NewSaptune(mockCommand)

	expectedDetails := Saptune{
		Version: "",
	}

	suite.EqualError(err, ErrSaptuneVersionUnknown.Error())
	suite.Equal(expectedDetails, saptuneRetriever)
}

func (suite *SaptuneTestSuite) TestNewSaptuneUnsupportedSaptuneVerErr() {
	mockCommand := new(mocks.CommandExecutor)

	mockCommand.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.0.0"), nil,
	)

	saptuneRetriever, _ := NewSaptune(mockCommand)
	statusOutput, err := saptuneRetriever.RunCommand("--format", "json", "status")

	expectedVersion := "3.0.0"
	expectedDetails := Saptune{
		Version:  expectedVersion,
		executor: mockCommand,
	}

	suite.EqualError(err, ErrUnsupportedSaptuneVer.Error())
	suite.Equal(expectedDetails, saptuneRetriever)
	suite.Equal([]byte(nil), statusOutput)
}
