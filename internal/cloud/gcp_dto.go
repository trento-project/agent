package cloud

import "strings"

type GcpMetadataDto struct {
	DiskNumber   int    `json:"disk_number"`
	Image        string `json:"image"`
	InstanceName string `json:"instance_name"`
	MachineType  string `json:"machine_type"`
	Network      string `json:"network"`
	ProjectID    string `json:"project_id"`
	Zone         string `json:"zone"`
}

func NewGcpMetadataDto(gcpMetadata *GcpMetadata) *GcpMetadataDto {
	return &GcpMetadataDto{
		DiskNumber:   len(gcpMetadata.Instance.Disks),
		Image:        lastSlashedString(gcpMetadata.Instance.Image),
		InstanceName: gcpMetadata.Instance.Name,
		MachineType:  lastSlashedString(gcpMetadata.Instance.MachineType),
		Network:      getNetwork(gcpMetadata),
		ProjectID:    gcpMetadata.Project.ProjectID,
		Zone:         lastSlashedString(gcpMetadata.Instance.Zone),
	}
}

func lastSlashedString(value string) string {
	splittedString := strings.Split(value, "/")
	return splittedString[len(splittedString)-1]
}

func getNetwork(gcpMetadata *GcpMetadata) string {
	var network string
	for _, val := range gcpMetadata.Instance.NetworkInterfaces {
		network = lastSlashedString(val.Network)
		break
	}
	return network
}
