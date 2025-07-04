package cloud_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cloud"
	"github.com/trento-project/agent/internal/core/cloud/mocks"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
)

type CloudMetadataTestSuite struct {
	suite.Suite
	mockExecutor   *utilsMocks.MockCommandExecutor
	mockHTTPClient *mocks.MockHTTPClient
}

func TestCloudMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(CloudMetadataTestSuite))
}

func (suite *CloudMetadataTestSuite) SetupTest() {
	suite.mockExecutor = new(utilsMocks.MockCommandExecutor)
	suite.mockHTTPClient = new(mocks.MockHTTPClient)
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

func systemdDetectVirtVmware() []byte {
	return []byte("vmware")
}

func systemdDetectVirtEmpty() []byte {
	return []byte("none")
}

func dmidecodeEmpty() []byte {
	return []byte("")
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderErr() {
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(nil, errors.New("error"))

	cIdentifier := cloud.NewIdentifier(suite.mockExecutor)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("", provider)
	suite.EqualError(err, "error")
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderAzure() {
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeAzure(), nil)

	cIdentifier := cloud.NewIdentifier(suite.mockExecutor)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("azure", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderAWSUsingSystemVersion() {
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-version").
		Return(dmidecodeAWSSystem(), nil)

	cIdentifier := cloud.NewIdentifier(suite.mockExecutor)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("aws", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderAWSUsingManufacturer() {
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-version").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-manufacturer").
		Return(dmidecodeAWSManufacturer(), nil)

	cIdentifier := cloud.NewIdentifier(suite.mockExecutor)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("aws", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderGCP() {
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-version").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-manufacturer").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "bios-vendor").
		Return(dmidecodeGCP(), nil)

	cIdentifier := cloud.NewIdentifier(suite.mockExecutor)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("gcp", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyProviderNutanix() {
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-version").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-manufacturer").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "bios-vendor").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode").
		Return(dmidecodeNutanix(), nil)

	cIdentifier := cloud.NewIdentifier(suite.mockExecutor)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("nutanix", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyProviderKVM() {
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-version").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-manufacturer").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "bios-vendor").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/bin/systemd-detect-virt").
		Return(systemdDetectVirtKVM(), nil)

	cIdentifier := cloud.NewIdentifier(suite.mockExecutor)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("kvm", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyProviderVmware() {
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-version").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-manufacturer").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "bios-vendor").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/bin/systemd-detect-virt").
		Return(systemdDetectVirtVmware(), nil)

	cIdentifier := cloud.NewIdentifier(suite.mockExecutor)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("vmware", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestIdentifyCloudProviderNoCloud() {
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-version").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-manufacturer").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "bios-vendor").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/bin/systemd-detect-virt").
		Return(systemdDetectVirtEmpty(), nil)

	cIdentifier := cloud.NewIdentifier(suite.mockExecutor)

	provider, err := cIdentifier.IdentifyCloudProvider()

	suite.Equal("", provider)
	suite.NoError(err)
}

func (suite *CloudMetadataTestSuite) TestNewCloudInstanceAzure() {
	ctx := context.TODO()

	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeAzure(), nil)

	body := io.NopCloser(bytes.NewReader([]byte(`{"compute":{"name":"test"}}`)))

	response := &http.Response{
		StatusCode: 200,
		Body:       body,
	}

	suite.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(
		response, nil,
	)

	c, err := cloud.NewCloudInstance(ctx, suite.mockExecutor, suite.mockHTTPClient)

	suite.NoError(err)
	suite.Equal("azure", c.Provider)
	meta, ok := c.Metadata.(*cloud.AzureMetadata)
	suite.True(ok)
	suite.Equal("test", meta.Compute.Name)
}

func (suite *CloudMetadataTestSuite) TestNewCloudInstanceAWS() {
	ctx := context.TODO()
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-version").
		Return(dmidecodeAWSSystem(), nil)

	tokenResponseBody := io.NopCloser(bytes.NewReader([]byte("token")))
	rootMetadataResponseBody := io.NopCloser(bytes.NewReader([]byte(`instance-id`)))
	instanceIDMetadataResponseBody := io.NopCloser(bytes.NewReader([]byte(`some-id`)))

	tokenResponse := &http.Response{
		StatusCode: 200,
		Body:       tokenResponseBody,
	}

	rootMetadataResponse := &http.Response{
		StatusCode: 200,
		Body:       rootMetadataResponseBody,
	}

	instanceIDMetadataResponse := &http.Response{
		StatusCode: 200,
		Body:       instanceIDMetadataResponseBody,
	}

	suite.mockHTTPClient.
		On("Do", mock.AnythingOfType("*http.Request")).
		Return(tokenResponse, nil).
		Once().
		On("Do", mock.AnythingOfType("*http.Request")).
		Return(rootMetadataResponse, nil).
		Once().
		On("Do", mock.AnythingOfType("*http.Request")).
		Return(instanceIDMetadataResponse, nil)

	c, err := cloud.NewCloudInstance(ctx, suite.mockExecutor, suite.mockHTTPClient)

	suite.NoError(err)
	suite.Equal("aws", c.Provider)
	meta, ok := c.Metadata.(*cloud.AWSMetadataDto)
	suite.True(ok)
	suite.Equal("some-id", meta.InstanceID)
}

func (suite *CloudMetadataTestSuite) TestNewAWSCloudInstanceWithoutMetadata() {
	ctx := context.TODO()
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-version").
		Return(dmidecodeAWSSystem(), nil)

	disabledIMDSResponse := &http.Response{
		StatusCode: 403,
		Body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}

	suite.mockHTTPClient.
		On("Do", mock.AnythingOfType("*http.Request")).
		Return(disabledIMDSResponse, nil).
		Once()

	c, err := cloud.NewCloudInstance(ctx, suite.mockExecutor, suite.mockHTTPClient)

	suite.NoError(err)
	suite.Equal("aws", c.Provider)
	suite.Nil(c.Metadata)
}

func (suite *CloudMetadataTestSuite) TestNewInstanceNutanix() {
	ctx := context.TODO()
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-version").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-manufacturer").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "bios-vendor").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode").
		Return(dmidecodeNutanix(), nil)

	c, err := cloud.NewCloudInstance(ctx, suite.mockExecutor, suite.mockHTTPClient)

	suite.NoError(err)
	suite.Equal("nutanix", c.Provider)
	suite.Equal(interface{}(nil), c.Metadata)
	suite.mockHTTPClient.AssertNotCalled(suite.T(), "Do")
}

func (suite *CloudMetadataTestSuite) TestNewInstanceKVM() {
	ctx := context.TODO()
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-version").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-manufacturer").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "bios-vendor").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/bin/systemd-detect-virt").
		Return(systemdDetectVirtKVM(), nil)

	c, err := cloud.NewCloudInstance(ctx, suite.mockExecutor, suite.mockHTTPClient)

	suite.NoError(err)
	suite.Equal("kvm", c.Provider)
	suite.Equal(interface{}(nil), c.Metadata)
	suite.mockHTTPClient.AssertNotCalled(suite.T(), "Do")
}

func (suite *CloudMetadataTestSuite) TestNewCloudInstanceNoCloud() {
	ctx := context.TODO()
	suite.mockExecutor.
		On("Exec", "/usr/sbin/dmidecode", "-s", "chassis-asset-tag").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-version").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "system-manufacturer").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode", "-s", "bios-vendor").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/sbin/dmidecode").
		Return(dmidecodeEmpty(), nil).
		On("Exec", "/usr/bin/systemd-detect-virt").
		Return(systemdDetectVirtEmpty(), nil)

	c, err := cloud.NewCloudInstance(ctx, suite.mockExecutor, suite.mockHTTPClient)

	suite.NoError(err)
	suite.Equal("", c.Provider)
	suite.Equal(interface{}(nil), c.Metadata)
	suite.mockHTTPClient.AssertNotCalled(suite.T(), "Do")
}
