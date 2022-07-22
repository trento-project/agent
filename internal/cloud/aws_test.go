package cloud

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trento-project/agent/internal/cloud/mocks"
)

func TestNewAwsMetadata(t *testing.T) {
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
		bodyText, _ := ioutil.ReadAll(aFile)
		body := ioutil.NopCloser(bytes.NewReader(bodyText))

		response := &http.Response{ //nolint
			StatusCode: 200,
			Body:       body,
		}

		clientMock.On("Do", mock.AnythingOfType("*http.Request")).Return(
			response, nil,
		).Once()
	}

	client = clientMock

	m, err := NewAwsMetadata()

	assert.NoError(t, err)

	assert.Equal(t, "some-ami-id", m.AmiID)
	assert.Equal(t, map[string]string{
		"root": "/dev/sda",
		"ebs1": "/dev/sdb1",
		"ebs2": "/dev/sdb2",
	}, m.BlockDeviceMapping)
	assert.Equal(t, "some-instance", m.InstanceID)
	assert.Equal(t, "some-instance-type", m.InstanceType)
	assert.Equal(t, "some-account-id", m.IdentityCredentials.EC2.Info.AccountID)
	assert.Equal(t, "some-vpc-id", m.Network.Interfaces.Macs["some-mac"].VpcID)
	assert.Equal(t, "some-availability-zone", m.Placement.AvailabilityZone)
	assert.Equal(t, "some-region", m.Placement.Region)
}
