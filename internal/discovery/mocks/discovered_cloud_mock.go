package mocks

import (
	"encoding/json"
	"io"
	"os"

	"github.com/trento-project/agent/internal/core/cloud"
	"github.com/trento-project/agent/internal/core/provider"
	"github.com/trento-project/agent/test/helpers"
)

func NewDiscoveredCloudMock() cloud.Instance {
	metadata := &cloud.AzureMetadata{} //nolint

	jsonFile, err := os.Open(helpers.GetFixturePath("discovery/azure/azure_metadata.json"))
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, metadata)

	if err != nil {
		panic(err)
	}
	return cloud.Instance{
		Provider: provider.Azure,
		Metadata: metadata,
	}
}
