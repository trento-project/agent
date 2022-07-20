package mocks

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/trento-project/agent/internal/cloud"
)

func NewDiscoveredCloudMock() cloud.Instance {
	metadata := &cloud.AzureMetadata{} //nolint

	jsonFile, err := os.Open("./test/fixtures/discovery/azure/azure_metadata.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, metadata)

	if err != nil {
		panic(err)
	}
	return cloud.Instance{
		Provider: cloud.Azure,
		Metadata: metadata,
	}
}
