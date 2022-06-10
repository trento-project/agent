package cloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAwsMetadataDto(t *testing.T) {

	awsMetadata := &AwsMetadata{
		AmiId: "some-ami",
		BlockDeviceMapping: map[string]string{
			"ami":  "sda",
			"ebs1": "/dev/sdb1",
			"ebs2": "/dev/sdb2",
			"root": "/dev/sda",
		},
		InstanceId:   "some-instance",
		InstanceType: "some-instance-type",
		Placement: Placement{
			AvailabilityZone: "some-availability-zone",
			Region:           "some-region",
		},
	}

	awsMetadata.IdentityCredentials.EC2.Info.AccountId = "some-account"
	awsMetadata.Network.Interfaces.Macs = make(map[string]MacEntry)
	macEntry := MacEntry{VpcId: "some-vpc-id"}
	awsMetadata.Network.Interfaces.Macs["eth1"] = macEntry

	awsMetadataDto := NewAwsMetadataDto(awsMetadata)

	expectedDto := &AwsMetadataDto{
		AccountId:        "some-account",
		AmiId:            "some-ami",
		AvailabilityZone: "some-availability-zone",
		DataDiskNumber:   2,
		InstanceId:       "some-instance",
		InstanceType:     "some-instance-type",
		Region:           "some-region",
		VpcId:            "some-vpc-id",
	}
	assert.Equal(t, expectedDto, awsMetadataDto)
}
