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

func NewGcpMetadataDto(GcpMetadata *GcpMetadata) *GcpMetadataDto {
	return &GcpMetadataDto{
		DiskNumber:   len(GcpMetadata.Instance.Disks),
		Image:        lastSlashedString(GcpMetadata.Instance.Image),
		InstanceName: GcpMetadata.Instance.Name,
		MachineType:  lastSlashedString(GcpMetadata.Instance.MachineType),
		Network:      getNetwork(GcpMetadata),
		ProjectID:    GcpMetadata.Project.ProjectID,
		Zone:         lastSlashedString(GcpMetadata.Instance.Zone),
	}
}

func lastSlashedString(value string) string {
	splittedString := strings.Split(value, "/")
	return splittedString[len(splittedString)-1]
}

func getNetwork(GcpMetadata *GcpMetadata) string {
	var network string
	for _, val := range GcpMetadata.Instance.NetworkInterfaces {
		network = lastSlashedString(val.Network)
		break
	}
	return network
}
