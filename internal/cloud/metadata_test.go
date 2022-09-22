package cloud

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/pkg/errors"
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

func dmidecodeAWSSystem() []byte {
	return []byte("4.11.amazon")
}

func dmidecodeAWSManufacturer() []byte {
	return []byte("Amazon EC2")
}

func dmidecodeGCP() []byte {
	return []byte("Google")
}

func dmidecodeNutanix() []byte {
	return []byte(`
		SomeUselessProp: some-value-1.1.0
		Version: nutanix-ahv-2.20220304.0.2429.el7
		Manufacturer: Nutanix
		Product Name: AHV
	`)
}

func systemdDetectVirtKVM() []byte {
	return []byte("kvm")
}

func systemdDetectVirtEmpty() []byte {
	return []byte("none")
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

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderAWSUsingSystemVersion() {
	mockCommand := new(utilsMocks.CommandExecutor)
	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-version").Return(
		dmidecodeAWSSystem(), nil,
	)

	cIdentifier := NewIdentifier(mockCommand)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("aws", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderAWSUsingManufacturer() {
	mockCommand := new(utilsMocks.CommandExecutor)

	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-version").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-manufacturer").Return(
		dmidecodeAWSManufacturer(), nil,
	)

	cIdentifier := NewIdentifier(mockCommand)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("aws", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderGCP() {
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
		dmidecodeGCP(), nil,
	)

	cIdentifier := NewIdentifier(mockCommand)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("gcp", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyProviderNutanix() {
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

	mockCommand.On("Exec", "dmidecode").Return(
		dmidecodeNutanix(), nil,
	)

	cIdentifier := NewIdentifier(mockCommand)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("nutanix", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyProviderKVM() {
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

	mockCommand.On("Exec", "dmidecode").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "systemd-detect-virt").Return(
		systemdDetectVirtKVM(), nil,
	)

	cIdentifier := NewIdentifier(mockCommand)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("kvm", provider)
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

	mockCommand.On("Exec", "dmidecode").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "systemd-detect-virt").Return(
		systemdDetectVirtEmpty(), nil,
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

func (suite *CloudMetadataTestSuite) TestNewCloudInstanceAWS() {
	mockCommand := new(utilsMocks.CommandExecutor)

	mockCommand.On("Exec", "dmidecode", "-s", "chassis-asset-tag").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "dmidecode", "-s", "system-version").Return(
		dmidecodeAWSSystem(), nil,
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
	meta, ok := c.Metadata.(*AWSMetadataDto)
	suite.True(ok)
	suite.Equal("some-id", meta.InstanceID)
}

func (suite *CloudMetadataTestSuite) TestNewInstanceNutanix() {
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

	mockCommand.On("Exec", "dmidecode").Return(
		dmidecodeNutanix(), nil,
	)

	c, err := NewCloudInstance(mockCommand)

	suite.NoError(err)
	suite.Equal("nutanix", c.Provider)
	suite.Equal(interface{}(nil), c.Metadata)
}

func (suite *CloudMetadataTestSuite) TestNewInstanceKVM() {
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

	mockCommand.On("Exec", "dmidecode").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "systemd-detect-virt").Return(
		systemdDetectVirtKVM(), nil,
	)

	c, err := NewCloudInstance(mockCommand)

	suite.NoError(err)
	suite.Equal("kvm", c.Provider)
	suite.Equal(interface{}(nil), c.Metadata)
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

	mockCommand.On("Exec", "dmidecode").Return(
		dmidecodeEmpty(), nil,
	)

	mockCommand.On("Exec", "systemd-detect-virt").Return(
		systemdDetectVirtEmpty(), nil,
	)

	c, err := NewCloudInstance(mockCommand)

	suite.NoError(err)
	suite.Equal("", c.Provider)
	suite.Equal(interface{}(nil), c.Metadata)
}
