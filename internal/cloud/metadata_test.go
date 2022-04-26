package cloud

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trento-project/agent/internal/cloud/mocks"
)

func mockDmidecodeErr() *exec.Cmd {
	return exec.Command("error")
}

func TestIdentifyCloudProviderErr(t *testing.T) {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeErr(),
	)

	provider, err := IdentifyCloudProvider()

	assert.Equal(t, "", provider)
	assert.EqualError(t, err, "exec: \"error\": executable file not found in $PATH")
}

func mockDmidecodeAzure() *exec.Cmd {
	return exec.Command("echo", "7783-7084-3265-9085-8269-3286-77")
}

func TestIdentifyCloudProviderAzure(t *testing.T) {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeAzure(),
	)

	provider, err := IdentifyCloudProvider()

	assert.Equal(t, "azure", provider)
	assert.NoError(t, err)
}

func mockDmidecodeAwsSystem() *exec.Cmd {
	return exec.Command("echo", "4.11.amazon")
}

func mockDmidecodeAwsManufacturer() *exec.Cmd {
	return exec.Command("echo", "Amazon EC2")
}

func TestIdentifyCloudProviderAwsUsingSystemVersion(t *testing.T) {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeNoCloud(),
	)

	mockCommand.On("Execute", "dmidecode", "-s", "system-version").Return(
		mockDmidecodeAwsSystem(),
	)

	provider, err := IdentifyCloudProvider()

	assert.Equal(t, "aws", provider)
	assert.NoError(t, err)
}

func TestIdentifyCloudProviderAwsUsingManufacturer(t *testing.T) {
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

	assert.Equal(t, "aws", provider)
	assert.NoError(t, err)
}

func mockDmidecodeGcp() *exec.Cmd {
	return exec.Command("echo", "Google")
}

func TestIdentifyCloudProviderGcp(t *testing.T) {
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

	assert.Equal(t, "gcp", provider)
	assert.NoError(t, err)
}

func mockDmidecodeNoCloud() *exec.Cmd {
	return exec.Command("echo", "")
}

func TestIdentifyCloudProviderNoCloud(t *testing.T) {
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

	assert.Equal(t, "", provider)
	assert.NoError(t, err)
}

func TestNewCloudInstanceAzure(t *testing.T) {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "dmidecode", "-s", "chassis-asset-tag").Return(
		mockDmidecodeAzure(),
	)

	clientMock := new(mocks.HTTPClient)

	body := ioutil.NopCloser(bytes.NewReader([]byte(`{"compute":{"name":"test"}}`)))

	response := &http.Response{
		StatusCode: 200,
		Body:       body,
	}

	clientMock.On("Do", mock.AnythingOfType("*http.Request")).Return(
		response, nil,
	)

	client = clientMock

	c, err := NewCloudInstance()

	assert.NoError(t, err)
	assert.Equal(t, "azure", c.Provider)
	meta := c.Metadata.(*AzureMetadata)
	assert.Equal(t, "test", meta.Compute.Name)
}

func TestNewCloudInstanceNoCloud(t *testing.T) {
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

	assert.NoError(t, err)
	assert.Equal(t, "", c.Provider)
	assert.Equal(t, interface{}(nil), c.Metadata)
}
