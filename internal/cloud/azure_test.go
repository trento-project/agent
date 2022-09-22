//nolint:lll
package cloud

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/cloud/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type AzureMetadataTestSuite struct {
	suite.Suite
}

func TestAzureMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(AzureMetadataTestSuite))
}

func (suite *AzureMetadataTestSuite) TestNewAzureMetadata() {
	clientMock := new(mocks.HTTPClient)

	aFile, _ := os.Open(helpers.GetFixtureFile("discovery/azure/azure_metadata.json"))
	bodyText, _ := io.ReadAll(aFile)
	body := io.NopCloser(bytes.NewReader(bodyText))

	response := &http.Response{ //nolint
		StatusCode: 200,
		Body:       body,
	}

	clientMock.On("Do", mock.AnythingOfType("*http.Request")).Return(
		response, nil,
	)

	client = clientMock

	m, err := NewAzureMetadata()

	expectedMeta := &AzureMetadata{
		Compute: Compute{
			AzEnvironment:              "AzurePublicCloud",
			EvictionPolicy:             "",
			IsHostCompatibilityLayerVM: "false",
			LicenseType:                "",
			Location:                   "westeurope",
			Name:                       "vmhana01",
			Offer:                      "sles-sap-15-sp2-byos",
			OsProfile: OsProfile{
				AdminUserName:                 "cloudadmin",
				ComputerName:                  "vmhana01",
				DisablePasswordAuthentication: "true",
			},
			OsType:           "Linux",
			PlacementGroupID: "",
			Plan: Plan{
				Name:      "",
				Product:   "",
				Publisher: "",
			},
			PlatformFaultDomain:  "1",
			PlatformUpdateDomain: "1",
			Priority:             "",
			Provider:             "Microsoft.Compute",

			PublicKeys: []*PublicKey{
				{
					KeyData: "ssh-rsa content\n",
					Path:    "/home/cloudadmin/.ssh/authorized_keys",
				},
			},
			Publisher:         "SUSE",
			ResourceGroupName: "test",
			ResourceID:        "/subscriptions/xxxxx/resourceGroups/test/providers/Microsoft.Compute/virtualMachines/vmhana01",
			SecurityProfile: SecurityProfile{
				SecureBootEnabled: "false",
				VirtualTpmEnabled: "false",
			},
			Sku: "gen2",
			StorageProfile: StorageProfile{
				DataDisks: []*Disk{
					{
						Caching:      "None",
						CreateOption: "Empty",
						DiskSizeGB:   "128",
						Image: map[string]string{
							"uri": "",
						},
						Lun: "0",
						ManagedDisk: ManagedDisk{
							ID:                 "/subscriptions/xxxxx/resourceGroups/test/providers/Microsoft.Compute/disks/disk-hana01-Data01", //nolint:lll
							StorageAccountType: "Premium_LRS",
						},
						Name: "disk-hana01-Data01",
						Vhd: map[string]string{
							"uri": "",
						},
						WriteAcceleratorEnabled: "false",
					},
					{
						Caching:      "None",
						CreateOption: "Empty",
						DiskSizeGB:   "128",
						Image: map[string]string{
							"uri": "",
						},
						Lun: "1",
						ManagedDisk: ManagedDisk{
							ID:                 "/subscriptions/xxxxx/resourceGroups/test/providers/Microsoft.Compute/disks/disk-hana01-Data02", //nolint:lll
							StorageAccountType: "Premium_LRS",
						},
						Name: "disk-hana01-Data02",
						Vhd: map[string]string{
							"uri": "",
						},
						WriteAcceleratorEnabled: "false",
					},
				},
				ImageReference: ImageReference{
					ID:        "",
					Offer:     "sles-sap-15-sp2-byos",
					Publisher: "SUSE",
					Sku:       "gen2",
					Version:   "latest",
				},
				OsDisk: Disk{
					Caching:      "ReadWrite",
					CreateOption: "FromImage",
					DiffDiskSettings: map[string]string{
						"option": "",
					},
					DiskSizeGB: "30",
					EncryptionSettings: map[string]string{
						"enabled": "false",
					},
					Image: map[string]string{
						"uri": "",
					},
					Lun: "",
					ManagedDisk: ManagedDisk{
						ID:                 "/subscriptions/xxxxx/resourceGroups/test/providers/Microsoft.Compute/disks/disk-hana01-Os",
						StorageAccountType: "Premium_LRS",
					},
					Name:   "disk-hana01-Os",
					OsType: "Linux",
					Vhd: map[string]string{
						"uri": "",
					},
					WriteAcceleratorEnabled: "false",
				},
			},
			SubscriptionID: "xxxxx",
			Tags:           "workspace:xdemo",
			TagsList: []map[string]string{
				{
					"name":  "workspace",
					"value": "xdemo",
				},
			},
			UserData:       "",
			Version:        "2021.06.05",
			VMID:           "data",
			VMScaleSetName: "",
			VMSize:         "Standard_E4s_v3",
			Zone:           "",
		},
		Network: Network{
			Interfaces: []*Interface{
				{
					Ipv4: IP{
						Addresses: []*Address{
							{
								PrivateIP: "10.74.1.10",
								PublicIP:  "1.2.3.4",
							},
						},
						Subnets: []*Subnet{
							{
								Address: "10.74.1.0",
								Prefix:  "24",
							},
						},
					},
					Ipv6: IP{
						Addresses: []*Address{},
						Subnets:   []*Subnet(nil),
					},
					MacAddress: "000D3A2267C3",
				},
			},
		},
	}

	suite.Equal(expectedMeta, m)
	suite.NoError(err)
}

func (suite *AzureMetadataTestSuite) TestGetVmUrl() {
	meta := &AzureMetadata{ //nolint
		Compute: Compute{ //nolint
			ResourceID: "myresourceid",
		},
	}

	suite.Equal("https:/portal.azure.com/#@SUSERDBillingsuse.onmicrosoft.com/resource/myresourceid", meta.GetVMURL())
}

func (suite *AzureMetadataTestSuite) TestGetResourceGroupUrl() {
	meta := &AzureMetadata{ //nolint
		Compute: Compute{ //nolint
			SubscriptionID:    "xxx",
			ResourceGroupName: "myresourcegroupname",
		},
	}

	suite.Equal("https:/portal.azure.com/#@SUSERDBillingsuse.onmicrosoft.com/resource/subscriptions/xxx/resourceGroups/myresourcegroupname/overview", meta.GetResourceGroupURL())
}
