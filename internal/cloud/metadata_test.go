package cloud

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/cloud/mocks"
	utilsMocks "github.com/trento-project/agent/internal/utils/mocks"
)

type CloudMetadataTestSuite struct {
	suite.Suite
}

func TestCloudMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(CloudMetadataTestSuite))
}

func dmidecodeAzure() []byte {
	return []byte("7783-7084-3265-9085-8269-3286-77")
}

func dmidecodeAwsSystem() []byte {
	return []byte("4.11.amazon")
}

func dmidecodeAwsManufacturer() []byte {
	return []byte("Amazon EC2")
}

func dmidecodeGcp() []byte {
	return []byte("Google")
}

func dmidecodeEmpty() []byte {
	return []byte("")
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderErr() {
	mockCommand := new(utilsMocks.CommandExecutor)
	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").Return(
		nil, errors.New("error"),
	)

	cIdentifier := NewIdentifier(mockCommand)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("", provider)
	suite.EqualError(err, "error")
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderAzure() {
	mockCommand := new(utilsMocks.CommandExecutor)
	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").Return(
		dmidecodeAzure(), nil,
	)

	cIdentifier := NewIdentifier(mockCommand)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("azure", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderAwsUsingSystemVersion() {
	mockCommand := new(utilsMocks.CommandExecutor)
	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-version").Return(
		dmidecodeAwsSystem(), nil,
	)

	cIdentifier := NewIdentifier(mockCommand)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("aws", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderAwsUsingManufacturer() {
	mockCommand := new(utilsMocks.CommandExecutor)

	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-version").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-manufacturer").Return(
		dmidecodeAwsManufacturer(), nil,
	)

	cIdentifier := NewIdentifier(mockCommand)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("aws", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderGcp() {
	mockCommand := new(utilsMocks.CommandExecutor)

	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-version").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-manufacturer").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "bios-vendor").Return(
		dmidecodeGcp(), nil,
	)

	cIdentifier := NewIdentifier(mockCommand)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("gcp", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderNoCloud() {
	mockCommand := new(utilsMocks.CommandExecutor)

	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-version").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-manufacturer").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "bios-vendor").Return(
		dmidecodeEmpty(), nil,
	)

	cIdentifier := NewIdentifier(mockCommand)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestNewCloudInstanceAzure() {
	mockCommand := new(utilsMocks.CommandExecutor)

	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").Return(
		dmidecodeAzure(), nil,
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

	c, err := NewCloudInstance(mockCommand)

	suite.NoError(err)
	suite.Equal("azure", c.Provider)
	meta, ok := c.Metadata.(*AzureMetadata)
	suite.True(ok)
	suite.Equal("test", meta.Compute.Name)
}

func (suite *CloudMetadataTestSuite) TestNewCloudInstanceAws() {
	mockCommand := new(utilsMocks.CommandExecutor)

	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-version").Return(
		dmidecodeAwsSystem(), nil,
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

	c, err := NewCloudInstance(mockCommand)

	suite.NoError(err)
	suite.Equal("aws", c.Provider)
	meta, ok := c.Metadata.(*AwsMetadataDto)
	suite.True(ok)
	suite.Equal("some-id", meta.InstanceID)
}

func (suite *CloudMetadataTestSuite) TestNewCloudInstanceNoCloud() {
	mockCommand := new(utilsMocks.CommandExecutor)

	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-version").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-manufacturer").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "bios-vendor").Return(
		dmidecodeEmpty(), nil,
	)

	c, err := NewCloudInstance(mockCommand)

	suite.NoError(err)
	suite.Equal("", c.Provider)
	suite.Equal(interface{}(nil), c.Metadata)
}
