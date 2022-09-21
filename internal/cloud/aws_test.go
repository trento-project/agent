package cloud

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/cloud/mocks"
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

	fixturesFolder := "../../test/fixtures/discovery/aws/"

	for _, fixture := range fixtures {
		aFile, _ := os.Open(path.Join(fixturesFolder, fixture))
		bodyText, _ := io.ReadAll(aFile)
		body := io.NopCloser(bytes.NewReader(bodyText))

		response := &http.Response{ //nolint
			StatusCode: 200,
			Body:       body,
		}

		clientMock.On("Do", mock.AnythingOfType("*http.Request")).Return(
			response, nil,
		).Once()
	}

	client = clientMock

	m, err := NewAWSMetadata()

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
