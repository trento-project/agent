package cloud

import (
	"bytes"
	"io"
	"net/http"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/cloud/mocks"
)

type CloudMetadataTestSuite struct {
	suite.Suite
}

func TestCloudMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(CloudMetadataTestSuite))
}

func mockDmidecodeErr() *exec.Cmd {
	return exec.Command("error")
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderErr() {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeErr(),
	)

	provider, err := IdentifyCloudProvider()

	suite.Equal("", provider)
	suite.EqualError(err, "exec: \"error\": executable file not found in $PATH")
}

func mockDmidecodeAzure() *exec.Cmd {
	return exec.Command("echo", "7783-7084-3265-9085-8269-3286-77")
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderAzure() {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeAzure(),
	)

	provider, err := IdentifyCloudProvider()

	suite.Equal("azure", provider)
	suite.NoError(err)
}

func mockDmidecodeAwsSystem() *exec.Cmd {
	return exec.Command("echo", "4.11.amazon")
}

func mockDmidecodeAwsManufacturer() *exec.Cmd {
	return exec.Command("echo", "Amazon EC2")
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderAwsUsingSystemVersion() {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "system-version").Return(
		mockDmidecodeAwsSystem(),
	)

	provider, err := IdentifyCloudProvider()

	suite.Equal("aws", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderAwsUsingManufacturer() {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "system-version").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "system-manufacturer").Return(
		mockDmidecodeAwsManufacturer(),
	)

	provider, err := IdentifyCloudProvider()

	suite.Equal("aws", provider)
	suite.NoError(err)
}

func mockDmidecodeGcp() *exec.Cmd {
	return exec.Command("echo", "Google")
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderGcp() {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "system-version").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "system-manufacturer").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "bios-vendor").Return(
		mockDmidecodeGcp(),
	)

	provider, err := IdentifyCloudProvider()

	suite.Equal("gcp", provider)
	suite.NoError(err)
}

func mockDmidecodeNoCloud() *exec.Cmd {
	return exec.Command("echo", "")
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderNoCloud() {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "system-version").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "system-manufacturer").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "bios-vendor").Return(
		mockDmidecodeNoCloud(),
	)

	provider, err := IdentifyCloudProvider()

	suite.Equal("", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestNewCloudInstanceAzure() {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeAzure(),
	)

	clientMock := new(mocks.HTTPClient)

	body := io.NopCloser(bytes.NewReader([]byte(`{"compute":{"name":"test"}}`)))

	response := &http.Response{ //nolint
		StatusCode: 200,
		Body:       body,
	}

	clientMock.On("Do", mock.AnythingOfType("*http.Request")).Return(
		response, nil,
	)

	client = clientMock

	c, err := NewCloudInstance()

	suite.NoError(err)
	suite.Equal("azure", c.Provider)
	meta, ok := c.Metadata.(*AzureMetadata)
	suite.True(ok)
	suite.Equal("test", meta.Compute.Name)
}

func (suite *CloudMetadataTestSuite) TestNewCloudInstanceAws() {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "system-version").Return(
		mockDmidecodeAwsSystem(),
	)

	clientMock := new(mocks.HTTPClient)

	request1 := io.NopCloser(bytes.NewReader([]byte(`instance-id`)))
	request2 := io.NopCloser(bytes.NewReader([]byte(`some-id`)))

	response1 := &http.Response{ //nolint
		StatusCode: 200,
		Body:       request1,
	}

	response2 := &http.Response{ //nolint
		StatusCode: 200,
		Body:       request2,
	}

	clientMock.On("Do", mock.AnythingOfType("*http.Request")).Return(
		response1, nil,
	).Once()

	clientMock.On("Do", mock.AnythingOfType("*http.Request")).Return(
		response2, nil,
	)

	client = clientMock

	c, err := NewCloudInstance()

	suite.NoError(err)
	suite.Equal("aws", c.Provider)
	meta, ok := c.Metadata.(*AwsMetadataDto)
	suite.True(ok)
	suite.Equal("some-id", meta.InstanceID)
}

func (suite *CloudMetadataTestSuite) TestNewCloudInstanceNoCloud() {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "system-version").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "system-manufacturer").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "bios-vendor").Return(
		mockDmidecodeNoCloud(),
	)

	c, err := NewCloudInstance()

	suite.NoError(err)
	suite.Equal("", c.Provider)
	suite.Equal(interface{}(nil), c.Metadata)
}
