/*
Based on
https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service?tabs=linux#instance-metadata
*/

package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	log "github.com/sirupsen/logrus"
)

const (
	azureAPIVersion = "2021-02-01"
	azureAPIAddress = "169.254.169.254"
	azurePortalURL  = "https://portal.azure.com/#@SUSERDBillingsuse.onmicrosoft.com/resource"
)

type AzureMetadata struct {
	Compute Compute `json:"compute,omitempty"`
	Network Network `json:"network,omitempty"`
}

type Compute struct {
	AzEnvironment              string              `json:"azEnvironment,omitempty"`
	EvictionPolicy             string              `json:"evictionPolicy,omitempty"`
	IsHostCompatibilityLayerVM string              `json:"isHostCompatibilityLayerVm,omitempty"`
	LicenseType                string              `json:"licenseType,omitempty"`
	Location                   string              `json:"location,omitempty"`
	Name                       string              `json:"name,omitempty"`
	Offer                      string              `json:"offer,omitempty"`
	OsProfile                  OsProfile           `json:"osProfile,omitempty"`
	OsType                     string              `json:"osType,omitempty"`
	PlacementGroupID           string              `json:"placementGroupId,omitempty"`
	Plan                       Plan                `json:"plan,omitempty"`
	PlatformFaultDomain        string              `json:"platformFaultDomain,omitempty"`
	PlatformUpdateDomain       string              `json:"platformUpdateDomain,omitempty"`
	Priority                   string              `json:"priority,omitempty"`
	Provider                   string              `json:"provider,omitempty"`
	PublicKeys                 []*PublicKey        `json:"publicKeys,omitempty"`
	Publisher                  string              `json:"publisher,omitempty"`
	ResourceGroupName          string              `json:"resourceGroupName,omitempty"`
	ResourceID                 string              `json:"resourceId,omitempty"`
	SecurityProfile            SecurityProfile     `json:"securityProfile,omitempty"`
	Sku                        string              `json:"sku,omitempty"`
	StorageProfile             StorageProfile      `json:"storageProfile,omitempty"`
	SubscriptionID             string              `json:"subscriptionId,omitempty"`
	Tags                       string              `json:"tags,omitempty"`
	TagsList                   []map[string]string `json:"tagsList,omitempty"`
	UserData                   string              `json:"userData,omitempty"`
	Version                    string              `json:"version,omitempty"`
	VMID                       string              `json:"vmId,omitempty"`
	VMScaleSetName             string              `json:"vmScaleSetName,omitempty"`
	VMSize                     string              `json:"vmSize,omitempty"`
	Zone                       string              `json:"zone,omitempty"`
}

type OsProfile struct {
	AdminUserName                 string `json:"adminUsername,omitempty"`
	ComputerName                  string `json:"computerName,omitempty"`
	DisablePasswordAuthentication string `json:"disablePasswordAuthentication,omitempty"`
}

type Plan struct {
	Name      string `json:"name,omitempty"`
	Product   string `json:"product,omitempty"`
	Publisher string `json:"publisher,omitempty"`
}

type PublicKey struct {
	KeyData string `json:"keyData,omitempty"`
	Path    string `json:"path,omitempty"`
}

type SecurityProfile struct {
	SecureBootEnabled string `json:"secureBootEnabled,omitempty"`
	VirtualTpmEnabled string `json:"virtualTpmEnabled,omitempty"`
}

type StorageProfile struct {
	DataDisks      []*Disk        `json:"dataDisks,omitempty"`
	ImageReference ImageReference `json:"imageReference,omitempty"`
	OsDisk         Disk           `json:"osDisk,omitempty"`
}

type Disk struct {
	Caching                 string            `json:"caching,omitempty"`
	CreateOption            string            `json:"createOption,omitempty"`
	DiffDiskSettings        map[string]string `json:"diffDiskSettings,omitempty"`
	DiskSizeGB              string            `json:"diskSizeGB,omitempty"`
	EncryptionSettings      map[string]string `json:"encryptionSettings,omitempty"`
	Image                   map[string]string `json:"image,omitempty"`
	Lun                     string            `json:"lun,omitempty"`
	ManagedDisk             ManagedDisk       `json:"managedDisk,omitempty"`
	Name                    string            `json:"name,omitempty"`
	OsType                  string            `json:"osType,omitempty"`
	Vhd                     map[string]string `json:"vhd,omitempty"`
	WriteAcceleratorEnabled string            `json:"writeAcceleratorEnabled,omitempty"`
}

type ManagedDisk struct {
	ID                 string `json:"id,omitempty"`
	StorageAccountType string `json:"storageAccountType,omitempty"`
}

type ImageReference struct {
	ID        string `json:"id,omitempty"`
	Offer     string `json:"offer,omitempty"`
	Publisher string `json:"publisher,omitempty"`
	Sku       string `json:"sku,omitempty"`
	Version   string `json:"version,omitempty"`
}

type Network struct {
	Interfaces []*Interface `json:"interface,omitempty"`
}

type Interface struct {
	Ipv4       IP     `json:"ipv4,omitempty"`
	Ipv6       IP     `json:"ipv6,omitempty"`
	MacAddress string `json:"macAddress,omitempty"`
}

type IP struct {
	Addresses []*Address `json:"ipAddress,omitempty"`
	Subnets   []*Subnet  `json:"subnet,omitempty"`
}

type Address struct {
	PrivateIP string `json:"privateIpAddress,omitempty"`
	PublicIP  string `json:"publicIpAddress,omitempty"`
}

type Subnet struct {
	Address string `json:"address,omitempty"`
	Prefix  string `json:"prefix,omitempty"`
}

func NewAzureMetadata(client HTTPClient) (*AzureMetadata, error) {
	var err error
	m := &AzureMetadata{
		Compute: Compute{
			AzEnvironment:              "",
			EvictionPolicy:             "",
			IsHostCompatibilityLayerVM: "",
			LicenseType:                "",
			Location:                   "",
			Name:                       "",
			Offer:                      "",
			OsProfile: OsProfile{
				AdminUserName:                 "",
				ComputerName:                  "",
				DisablePasswordAuthentication: "",
			},
			OsType:           "",
			PlacementGroupID: "",
			Plan: Plan{
				Name:      "",
				Product:   "",
				Publisher: "",
			},
			PlatformFaultDomain:  "",
			PlatformUpdateDomain: "",
			Priority:             "",
			Provider:             "",
			PublicKeys:           []*PublicKey{},
			Publisher:            "",
			ResourceGroupName:    "",
			ResourceID:           "",
			SecurityProfile: SecurityProfile{
				SecureBootEnabled: "",
				VirtualTpmEnabled: "",
			},
			Sku: "",
			StorageProfile: StorageProfile{
				DataDisks: []*Disk{},
				ImageReference: ImageReference{
					ID:        "",
					Offer:     "",
					Publisher: "",
					Sku:       "",
					Version:   "",
				},
				OsDisk: Disk{
					Caching:            "",
					CreateOption:       "",
					DiffDiskSettings:   map[string]string{},
					DiskSizeGB:         "",
					EncryptionSettings: map[string]string{},
					Image:              map[string]string{},
					Lun:                "",
					ManagedDisk: ManagedDisk{
						ID:                 "",
						StorageAccountType: "",
					},
					Name:                    "",
					OsType:                  "",
					Vhd:                     map[string]string{},
					WriteAcceleratorEnabled: "",
				},
			},
			SubscriptionID: "",
			Tags:           "",
			TagsList:       []map[string]string{},
			UserData:       "",
			Version:        "",
			VMID:           "",
			VMScaleSetName: "",
			VMSize:         "",
			Zone:           "",
		},
		Network: Network{
			Interfaces: []*Interface{},
		},
	}

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/metadata/instance", azureAPIAddress), nil)
	req.Header.Add("Metadata", "True")

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("api-version", azureAPIVersion)
	req.URL.RawQuery = q.Encode()

	log.Debug("Requesting Azure metadata...")

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var pjson bytes.Buffer
	err = json.Indent(&pjson, body, "", " ")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Debugln(pjson.String())

	err = json.Unmarshal(body, m)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return m, nil
}

func (m *AzureMetadata) GetVMURL() string {
	return path.Join(azurePortalURL, m.Compute.ResourceID)
}

func (m *AzureMetadata) GetResourceGroupURL() string {
	return path.Join(
		azurePortalURL,
		"subscriptions",
		m.Compute.SubscriptionID,
		"resourceGroups",
		m.Compute.ResourceGroupName,
		"overview",
	)
}
