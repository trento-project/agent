package cloud

type AwsMetadataDto struct {
	AccountId        string `json:"account_id"`
	AmiId            string `json:"ami_id"`
	AvailabilityZone string `json:"availability_zone"`
	DataDiskNumber   int    `json:"data_disk_number"`
	InstanceId       string `json:"instance_id"`
	InstanceType     string `json:"instance_type"`
	Region           string `json:"region"`
	VpcId            string `json:"vpc_id"`
}

func NewAwsMetadataDto(awsMetadata *AwsMetadata) *AwsMetadataDto {
	return &AwsMetadataDto{
		AccountId:        awsMetadata.IdentityCredentials.EC2.Info.AccountId,
		AmiId:            awsMetadata.AmiId,
		AvailabilityZone: awsMetadata.Placement.AvailabilityZone,
		DataDiskNumber:   getDataDiskNumber(awsMetadata),
		InstanceId:       awsMetadata.InstanceId,
		InstanceType:     awsMetadata.InstanceType,
		Region:           awsMetadata.Placement.Region,
		VpcId:            getVpcId(awsMetadata),
	}
}

func getDataDiskNumber(awsMetadata *AwsMetadata) int {
	var dataDiskNumber int = 0
	for device := range awsMetadata.BlockDeviceMapping {
		if device != "root" && device != "ami" {
			dataDiskNumber++
		}
	}

	return dataDiskNumber
}

func getVpcId(awsMetadata *AwsMetadata) string {
	var vpcId string
	for _, val := range awsMetadata.Network.Interfaces.Macs {
		vpcId = val.VpcId
		break
	}
	return vpcId
}
