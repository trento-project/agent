package cloud_test

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cloud"
	"github.com/trento-project/agent/internal/core/cloud/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type AWSMetadataTestSuite struct {
	suite.Suite
}

func TestAWSMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(AWSMetadataTestSuite))
}

func (suite *AWSMetadataTestSuite) TestNewAWSMetadata() {
	clientMock := new(mocks.HTTPClient)

	fixtures := []string{
		"meta-data",
		"ami-id",
		"block-device-mapping",
		"ebs1",
		"ebs2",
		"root",
		"identity-credentials",
		"ec2",
		"info",
		"instance-id",
		"instance-type",
		"network",
		"interfaces",
		"macs",
		"some-mac",
		"vpc-id",
		"placement",
		"availability-zone",
		"region",
	}

	fixturesFolder := helpers.GetFixturePath("discovery/aws")

	for _, fixture := range fixtures {
		aFile, _ := os.Open(path.Join(fixturesFolder, fixture))
		bodyText, _ := io.ReadAll(aFile)
		body := io.NopCloser(bytes.NewReader(bodyText))

		response := &http.Response{
			StatusCode: 200,
			Body:       body,
		}

		clientMock.On("Do", mock.AnythingOfType("*http.Request")).Return(
			response, nil,
		).Once()
	}

	m, err := cloud.NewAWSMetadata(clientMock)

	suite.NoError(err)

	suite.Equal("some-ami-id", m.AmiID)
	suite.Equal(map[string]string{
		"root": "/dev/sda",
		"ebs1": "/dev/sdb1",
		"ebs2": "/dev/sdb2",
	}, m.BlockDeviceMapping)
	suite.Equal("some-instance", m.InstanceID)
	suite.Equal("some-instance-type", m.InstanceType)
	suite.Equal("some-account-id", m.IdentityCredentials.EC2.Info.AccountID)
	suite.Equal("some-vpc-id", m.Network.Interfaces.Macs["some-mac"].VpcID)
	suite.Equal("some-availability-zone", m.Placement.AvailabilityZone)
	suite.Equal("some-region", m.Placement.Region)
}
