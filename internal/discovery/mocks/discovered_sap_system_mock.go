package mocks

import (
	"encoding/json"
	"io"
	"os"

	"github.com/trento-project/agent/internal/sapsystem"
	"github.com/trento-project/agent/test/helpers"
)

func NewDiscoveredSAPSystemDatabaseMock() sapsystem.SAPSystemsList {
	var s sapsystem.SAPSystemsList

	jsonFile, err := os.Open(helpers.GetFixturePath("discovery/sap_system/sap_system_discovery_database.json"))
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &s)

	if err != nil {
		panic(err)
	}
	return s
}

func NewDiscoveredSAPSystemApplicationMock() sapsystem.SAPSystemsList {
	var s sapsystem.SAPSystemsList

	jsonFile, err := os.Open(helpers.GetFixturePath("discovery/sap_system/sap_system_discovery_application.json"))
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &s)
	if err != nil {
		panic(err)
	}
	return s
}

func NewDiscoveredSAPSystemDiagnosticsMock() sapsystem.SAPSystemsList {
	var s sapsystem.SAPSystemsList

	jsonFile, err := os.Open(helpers.GetFixturePath("discovery/sap_system/sap_system_discovery_diagnostics.json"))
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &s)
	if err != nil {
		panic(err)
	}
	return s
}
