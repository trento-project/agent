package cloud_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cloud"
	"github.com/trento-project/agent/internal/core/cloud/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type AzureMetadataTestSuite struct {
	suite.Suite
}

func TestAzureMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(AzureMetadataTestSuite))
}

func (suite *AzureMetadataTestSuite) TestNewAzureMetadata() {
	ctx := context.TODO()
	clientMock := new(mocks.MockHTTPClient)

	aFile, _ := os.Open(helpers.GetFixturePath("discovery/azure/azure_metadata.json"))
	bodyText, _ := io.ReadAll(aFile)
	body := io.NopCloser(bytes.NewReader(bodyText))

	response := &http.Response{
		StatusCode: 200,
		Body:       body,
	}

	clientMock.On("Do", mock.AnythingOfType("*http.Request")).Return(
		response, nil,
	)

	m, err := cloud.NewAzureMetadata(ctx, clientMock)

	expectedMeta := &cloud.AzureMetadata{
		Compute: cloud.Compute{
			AzEnvironment:              "AzurePublicCloud",
			EvictionPolicy:             "",
			IsHostCompatibilityLayerVM: "false",
			LicenseType:                "",
			Location:                   "westeurope",
			Name:                       "vmhana01",
			Offer:                      "sles-sap-15-sp2-byos",
			OsProfile: cloud.OsProfile{
				AdminUserName:                 "cloudadmin",
				ComputerName:                  "vmhana01",
				DisablePasswordAuthentication: "true",
			},
			OsType:           "Linux",
			PlacementGroupID: "",
			Plan: cloud.Plan{
				Name:      "",
				Product:   "",
				Publisher: "",
			},
			PlatformFaultDomain:  "1",
			PlatformUpdateDomain: "1",
			Priority:             "",
			Provider:             "Microsoft.Compute",

			PublicKeys: []*cloud.PublicKey{
				{
					KeyData: "ssh-rsa content\n",
					Path:    "/home/cloudadmin/.ssh/authorized_keys",
				},
			},
			Publisher:         "SUSE",
			ResourceGroupName: "test",
			ResourceID:        "/subscriptions/xxxxx/resourceGroups/test/providers/Microsoft.Compute/virtualMachines/vmhana01",
			SecurityProfile: cloud.SecurityProfile{
				SecureBootEnabled: "false",
				VirtualTpmEnabled: "false",
			},
			Sku: "gen2",
			StorageProfile: cloud.StorageProfile{
				DataDisks: []*cloud.Disk{
					{
						Caching:      "None",
						CreateOption: "Empty",
						DiskSizeGB:   "128",
						Image: map[string]string{
							"uri": "",
						},
						Lun: "0",
						ManagedDisk: cloud.ManagedDisk{
							ID:                 "/subscriptions/xxxxx/resourceGroups/test/providers/Microsoft.Compute/disks/disk-hana01-Data01",
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
						ManagedDisk: cloud.ManagedDisk{
							ID:                 "/subscriptions/xxxxx/resourceGroups/test/providers/Microsoft.Compute/disks/disk-hana01-Data02",
							StorageAccountType: "Premium_LRS",
						},
						Name: "disk-hana01-Data02",
						Vhd: map[string]string{
							"uri": "",
						},
						WriteAcceleratorEnabled: "false",
					},
				},
				ImageReference: cloud.ImageReference{
					ID:        "",
					Offer:     "sles-sap-15-sp2-byos",
					Publisher: "SUSE",
					Sku:       "gen2",
					Version:   "latest",
				},
				OsDisk: cloud.Disk{
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
					ManagedDisk: cloud.ManagedDisk{
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
		Network: cloud.Network{
			Interfaces: []*cloud.Interface{
				{
					Ipv4: cloud.IP{
						Addresses: []*cloud.Address{
							{
								PrivateIP: "10.74.1.10",
								PublicIP:  "1.2.3.4",
							},
						},
						Subnets: []*cloud.Subnet{
							{
								Address: "10.74.1.0",
								Prefix:  "24",
							},
						},
					},
					Ipv6: cloud.IP{
						Addresses: []*cloud.Address{},
						Subnets:   []*cloud.Subnet(nil),
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
	meta := &cloud.AzureMetadata{
		Compute: cloud.Compute{
			ResourceID: "myresourceid",
		},
	}

	suite.Equal("https:/portal.azure.com/#@SUSERDBillingsuse.onmicrosoft.com/resource/myresourceid", meta.GetVMURL())
}

func (suite *AzureMetadataTestSuite) TestGetResourceGroupUrl() {
	meta := &cloud.AzureMetadata{
		Compute: cloud.Compute{
			SubscriptionID:    "xxx",
			ResourceGroupName: "myresourcegroupname",
		},
	}

	suite.Equal("https:/portal.azure.com/#@SUSERDBillingsuse.onmicrosoft.com/resource/subscriptions/xxx/resourceGroups/myresourcegroupname/overview", meta.GetResourceGroupURL())
}
