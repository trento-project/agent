package cloud

type AwsMetadataDto struct {
	AccountID        string `json:"account_id"`
	AmiID            string `json:"ami_id"`
	AvailabilityZone string `json:"availability_zone"`
	DataDiskNumber   int    `json:"data_disk_number"`
	InstanceID       string `json:"instance_id"`
	InstanceType     string `json:"instance_type"`
	Region           string `json:"region"`
	VpcID            string `json:"vpc_id"`
}

func NewAwsMetadataDto(awsMetadata *AwsMetadata) *AwsMetadataDto {
	return &AwsMetadataDto{
		AccountID:        awsMetadata.IdentityCredentials.EC2.Info.AccountID,
		AmiID:            awsMetadata.AmiID,
		AvailabilityZone: awsMetadata.Placement.AvailabilityZone,
		DataDiskNumber:   getDataDiskNumber(awsMetadata),
		InstanceID:       awsMetadata.InstanceID,
		InstanceType:     awsMetadata.InstanceType,
		Region:           awsMetadata.Placement.Region,
		VpcID:            getVpcID(awsMetadata),
	}
}

func getDataDiskNumber(awsMetadata *AwsMetadata) int {
	var dataDiskNumber int
	for device := range awsMetadata.BlockDeviceMapping {
		if device != "root" && device != "ami" {
			dataDiskNumber++
		}
	}

	return dataDiskNumber
}

func getVpcID(awsMetadata *AwsMetadata) string {
	var vpcID string
	for _, val := range awsMetadata.Network.Interfaces.Macs {
		vpcID = val.VpcID
		break
	}
	return vpcID
}
