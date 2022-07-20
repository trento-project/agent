package cloud

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trento-project/agent/internal/cloud/mocks"
)

func TestNewGcpMetadata(t *testing.T) {
	clientMock := new(mocks.HTTPClient)

	aFile, _ := os.Open("../../test/fixtures/discovery/gcp/gcp_metadata.json")
	bodyText, _ := ioutil.ReadAll(aFile)
	body := ioutil.NopCloser(bytes.NewReader(bodyText))

	response := &http.Response{ //nolint
		StatusCode: 200,
		Body:       body,
	}

	clientMock.On("Do", mock.AnythingOfType("*http.Request")).Return(
		response, nil,
	)

	client = clientMock

	m, err := NewGcpMetadata()

	expectedMeta := &GcpMetadata{
		Instance: GcpInstance{
			Disks: []GcpDisk{
				{
					DeviceName: "persistent-disk-0",
					Index:      0,
				},
				{
					DeviceName: "hana-data-0",
					Index:      1,
				},
				{
					DeviceName: "hana-backup-0",
					Index:      2,
				},
				{
					DeviceName: "hana-software-0",
					Index:      3,
				},
			},
			Image:       "projects/suse-byos-cloud/global/images/sles-15-sp1-sap-byos-v20220126",
			MachineType: "projects/123456/machineTypes/n1-highmem-8",
			Name:        "vmhana01",
			NetworkInterfaces: []GcpNetworkInterface{
				{
					Network: "projects/123456/networks/network",
				},
			},
			Zone: "projects/123456/zones/europe-west1-b",
		},
		Project: GcpProject{
			ProjectID: "some-project-id",
		},
	}

	assert.Equal(t, expectedMeta, m)
	assert.NoError(t, err)
}
