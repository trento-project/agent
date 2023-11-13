package cloud_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cloud"
)

type GCPMetadataDtoTestSuite struct {
	suite.Suite
}

func TestGCPMetadataDtoTestSuite(t *testing.T) {
	suite.Run(t, new(GCPMetadataDtoTestSuite))
}

func (suite *GCPMetadataDtoTestSuite) TestNewGCPMetadataDto() {

	gcpMetadata := &cloud.GCPMetadata{
		Instance: cloud.GCPInstance{
			Disks: []cloud.GCPDisk{
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
			NetworkInterfaces: []cloud.GCPNetworkInterface{
				{
					Network: "projects/123456/networks/network",
				},
			},
			Zone: "projects/123456/zones/europe-west1-b",
		},
		Project: cloud.GCPProject{
			ProjectID: "some-project-id",
		},
	}

	gcpMetadataDto := cloud.NewGCPMetadataDto(gcpMetadata)

	expectedDto := &cloud.GCPMetadataDto{
		DiskNumber:   4,
		Image:        "sles-15-sp1-sap-byos-v20220126",
		InstanceName: "vmhana01",
		MachineType:  "n1-highmem-8",
		Network:      "network",
		ProjectID:    "some-project-id",
		Zone:         "europe-west1-b",
	}
	suite.Equal(expectedDto, gcpMetadataDto)
}
