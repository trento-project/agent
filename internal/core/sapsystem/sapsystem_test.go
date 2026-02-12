//nolint:nosnakecase,dupl
package sapsystem_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/sapsystem"
	"github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi"
	sapControlMocks "github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi/mocks"
	"github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type SAPSystemTestSuite struct {
	suite.Suite
}

func TestSAPSystemTestSuite(t *testing.T) {
	testSuite := new(SAPSystemTestSuite)
	suite.Run(t, testSuite)
}

func fakeNewWebService(instName string, features string) sapcontrolapi.WebService {
	mockWebService := new(sapControlMocks.MockWebService)
	ctx := context.TODO()
	mockWebService.
		On("GetInstancePropertiesContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetInstancePropertiesResponse{
			Properties: []*sapcontrolapi.InstanceProperty{
				{
					Property:     "SAPSYSTEMNAME",
					Propertytype: "string",
					Value:        "DEV",
				},
				{
					Property:     "INSTANCE_NAME",
					Propertytype: "string",
					Value:        instName,
				},
				{
					Property:     "SAPLOCALHOST",
					Propertytype: "string",
					Value:        "host",
				},
				{
					Property:     "SAPSYSTEM",
					Propertytype: "string",
					Value:        instName[len(instName)-2:],
				},
			},
		}, nil)

	mockWebService.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetProcessListResponse{
			Processes: []*sapcontrolapi.OSProcess{},
		}, nil)

	mockWebService.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetSystemInstanceListResponse{
			Instances: []*sapcontrolapi.SAPInstance{
				{
					Hostname:      "host",
					InstanceNr:    0,
					HttpPort:      50013,
					HttpsPort:     50014,
					StartPriority: "0.3",
					Features:      features,
					Dispstatus:    sapcontrolapi.STATECOLOR_GREEN,
				},
			},
		}, nil)

	return mockWebService
}

func mockDEVFileSystem() (afero.Fs, error) {
	appFS := afero.NewMemMapFs()
	err := appFS.MkdirAll("/usr/sap/DEV/ASCS01", 0755)
	if err != nil {
		return nil, err
	}
	err = afero.WriteFile(appFS, "/usr/sap/DEV/SYS/profile/DEFAULT.PFL", []byte{}, 0644)
	if err != nil {
		return nil, err
	}
	err = appFS.MkdirAll("/usr/sap/DEV/SYS/global/hdb/custom/config/", 0755)
	if err != nil {
		return nil, err
	}
	err = appFS.MkdirAll("/usr/sap/DEV/SYS/global/sapcontrol", 0755)
	if err != nil {
		return nil, err
	}
	err = afero.WriteFile(
		appFS,
		"/usr/sap/DEV/SYS/global/sapcontrol/0.3_50013_50014_0_2_00_host",
		[]byte("Host:somehost Pid:100"),
		0644,
	)
	if err != nil {
		return nil, err
	}
	return appFS, nil
}

func mockLandscapeHostConfiguration() []byte {
	lFile, err := os.Open(helpers.GetFixturePath("discovery/sap_system/landscape_host_configuration"))
	if err != nil {
		panic(err)
	}
	content, err := io.ReadAll(lFile)
	if err != nil {
		panic(err)
	}
	return content
}

func mockHdbnsutilSrstate() []byte {
	lFile, err := os.Open(helpers.GetFixturePath("discovery/sap_system/hdbnsutil_srstate"))
	if err != nil {
		panic(err)
	}
	content, err := io.ReadAll(lFile)
	if err != nil {
		panic(err)
	}
	return content
}

func mockSappfpar() []byte {
	return []byte("systemId")
}

func (suite *SAPSystemTestSuite) TestNewSAPSystemsList() {
	ctx := context.TODO()
	appFS := afero.NewMemMapFs()
	err := appFS.MkdirAll("/usr/sap/DEV/ASCS01", 0755)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/usr/sap/DEV/SYS/profile/DEFAULT.PFL", []byte{}, 0644)
	suite.NoError(err)
	err = appFS.MkdirAll("/usr/sap/PRD/ERS02", 0755)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/usr/sap/PRD/SYS/profile/DEFAULT.PFL", []byte{}, 0644)
	suite.NoError(err)

	mockCommand := new(mocks.MockCommandExecutor)
	mockWebServiceConnector := new(sapControlMocks.MockWebServiceConnector)

	mockWebServiceConnector.On("New", "01").Return(fakeNewWebService("ASCS01", ""))
	mockWebServiceConnector.On("New", "02").Return(fakeNewWebService("ERS02", ""))

	systems, err := sapsystem.NewSAPSystemsList(ctx, appFS, mockCommand, mockWebServiceConnector)

	suite.Len(systems, 2)
	suite.Equal(systems[0].SID, "DEV")
	suite.Equal(systems[1].SID, "PRD")
	suite.NoError(err)
}

func (suite *SAPSystemTestSuite) TestNewSAPSystem() {
	ctx := context.TODO()
	mockCommand := new(mocks.MockCommandExecutor)
	mockWebServiceConnector := new(sapControlMocks.MockWebServiceConnector)
	mockWebServiceConnector.On("New", "01").Return(fakeNewWebService("ASCS01", ""))
	mockWebServiceConnector.On("New", "02").Return(fakeNewWebService("ERS02", ""))

	appFS := afero.NewMemMapFs()
	err := appFS.MkdirAll("/usr/sap/DEV/ASCS01", 0755)
	suite.NoError(err)
	err = appFS.MkdirAll("/usr/sap/DEV/ERS02", 0755)
	suite.NoError(err)

	profileFile, _ := os.Open(helpers.GetFixturePath("discovery/sap_system/sap_profile_default"))
	profileContent, _ := io.ReadAll(profileFile)

	err = appFS.MkdirAll("/usr/sap/DEV/SYS/profile", 0755)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/usr/sap/DEV/SYS/profile/DEFAULT.PFL", profileContent, 0644)
	suite.NoError(err)

	err = appFS.MkdirAll("/usr/sap/DEV/SYS/global/sapcontrol", 0755)
	suite.NoError(err)

	expectedProfile := sapsystem.SAPProfile{
		"SAPSYSTEMNAME":                "HA1",
		"SAPGLOBALHOST":                "sapha1as",
		"rdisp/mshost":                 "sapha1as",
		"rdisp/msserv":                 "sapmsHA1",
		"rdisp/msserv_internal":        "3900",
		"enque/process_location":       "REMOTESA",
		"enque/serverhost":             "sapha1as",
		"enque/serverinst":             "00",
		"is/HTTP/show_detailed_errors": "FALSE",
		"icf/user_recheck":             "1",
		"icm/HTTP/ASJava/disable_url_session_tracking": "TRUE",
		"service/protectedwebmethods":                  "SDEFAULT",
		"rsec/ssfs_datapath":                           "$(DIR_GLOBAL)$(DIR_SEP)security$(DIR_SEP)rsecssfs$(DIR_SEP)data",
		"rsec/ssfs_keypath":                            "$(DIR_GLOBAL)$(DIR_SEP)security$(DIR_SEP)rsecssfs$(DIR_SEP)key",
		"gw/sec_info":                                  "$(DIR_GLOBAL)$(DIR_SEP)secinfo$(FT_DAT)",
		"login/system_client":                          "001",
		"enque/deque_wait_answer":                      "TRUE",
		"system/type":                                  "ABAP",
		"SAPDBHOST":                                    "192.168.140.12",
		"j2ee/dbtype":                                  "hdb",
		"j2ee/dbname":                                  "PRD",
		"j2ee/dbhost":                                  "192.168.140.12",
		"dbs/hdb/dbname":                               "PRD",
		"rsdb/ssfs_connect":                            "0",
		"dbs/hdb/schema":                               "SAPABAP1",
		"gw/acl_mode":                                  "1",
		"login/password_downwards_compatibility":       "0",
		"vmcj/enable":                                  "off",
	}

	sappfparCmd := "sappfpar SAPSYSTEMNAME SAPGLOBALHOST SAPFQDN SAPDBHOST dbs/hdb/dbname dbs/hdb/schema rdisp/msp/msserv rdisp/msserv_internal name=DEV"
	mockCommand.On("Output", "/usr/bin/su", "-lc", sappfparCmd, "devadm").Return(mockSappfpar(), nil)

	system, err := sapsystem.NewSAPSystem(ctx, appFS, mockCommand, mockWebServiceConnector, "/usr/sap/DEV")

	suite.Equal(sapsystem.Unknown, system.Type)
	suite.Contains("ASCS01", system.Instances[0].Name)
	suite.Contains("ERS02", system.Instances[1].Name)
	suite.Equal(expectedProfile, system.Profile)
	suite.NoError(err)
}

func mockSystemReplicationStatus() []byte {
	sFile, err := os.Open(helpers.GetFixturePath("discovery/sap_system/system_replication_status"))
	if err != nil {
		panic(err)
	}
	content, err := io.ReadAll(sFile)
	if err != nil {
		panic(err)
	}
	return content
}

func (suite *SAPSystemTestSuite) TestDetectSystemId_Database() {
	ctx := context.TODO()
	appFS, err := mockDEVFileSystem()
	suite.NoError(err)
	nameserverContent := []byte(`
key1 = value1
id = systemId
key2 = value2
`)

	err = afero.WriteFile(
		appFS, "/usr/sap/DEV/SYS/global/hdb/custom/config/nameserver.ini",
		nameserverContent, 0644)
	suite.NoError(err)

	mockCommand := new(mocks.MockCommandExecutor)
	mockWebServiceConnector := new(sapControlMocks.MockWebServiceConnector)

	mockWebServiceConnector.On("New", "01").Return(fakeNewWebService("HDB00", "HDB"))
	mockCommand.
		On("Output", "/usr/bin/su", "-lc", "python /usr/sap/DEV/HDB00/exe/python_support/systemReplicationStatus.py --sapcontrol=1", "devadm").
		Return(mockSystemReplicationStatus(), nil).
		On("Output", "/usr/bin/su", "-lc", "python /usr/sap/DEV/HDB00/exe/python_support/landscapeHostConfiguration.py --sapcontrol=1", "devadm").
		Return(mockLandscapeHostConfiguration(), nil).
		On("Output", "/usr/bin/su", "-lc", "/usr/sap/DEV/HDB00/exe/hdbnsutil -sr_state -sapcontrol=1", "devadm").
		Return(mockHdbnsutilSrstate(), nil)

	system, err := sapsystem.NewSAPSystem(ctx, appFS, mockCommand, mockWebServiceConnector, "/usr/sap/DEV")

	suite.Equal("089d1a278481b86e821237f8e98e6de7", system.ID)
	suite.Equal(sapsystem.Database, system.Type)
	suite.NoError(err)
}

func (suite *SAPSystemTestSuite) TestDetectSystemId_Application() {
	ctx := context.TODO()
	appFS, err := mockDEVFileSystem()
	suite.NoError(err)
	mockCommand := new(mocks.MockCommandExecutor)
	mockWebServiceConnector := new(sapControlMocks.MockWebServiceConnector)

	mockWebServiceConnector.On("New", "01").Return(fakeNewWebService("HDB00", "MESSAGESERVER|ENQUE"))
	sappfparCmd := "sappfpar SAPSYSTEMNAME SAPGLOBALHOST SAPFQDN SAPDBHOST dbs/hdb/dbname dbs/hdb/schema rdisp/msp/msserv rdisp/msserv_internal name=DEV"
	mockCommand.On("Output", "/usr/bin/su", "-lc", sappfparCmd, "devadm").Return(mockSappfpar(), nil)

	system, err := sapsystem.NewSAPSystem(ctx, appFS, mockCommand, mockWebServiceConnector, "/usr/sap/DEV")

	suite.Equal("089d1a278481b86e821237f8e98e6de7", system.ID)
	suite.Equal(sapsystem.Application, system.Type)
	suite.NoError(err)
}

func (suite *SAPSystemTestSuite) TestDetectSystemId_Diagnostics() {
	ctx := context.TODO()
	appFS, err := mockDEVFileSystem()
	suite.NoError(err)
	machineIDContent := []byte(`dummy-machine-id`)

	err = afero.WriteFile(
		appFS, "/etc/machine-id",
		machineIDContent, 0644)
	suite.NoError(err)

	mockCommand := new(mocks.MockCommandExecutor)
	mockWebServiceConnector := new(sapControlMocks.MockWebServiceConnector)

	mockWebServiceConnector.On("New", "01").Return(fakeNewWebService("HDB00", "SMDAGENT"))

	system, err := sapsystem.NewSAPSystem(ctx, appFS, mockCommand, mockWebServiceConnector, "/usr/sap/DEV")

	suite.Equal("d3d5dd5ec501127e0011a2531e3b11ff", system.ID)
	suite.Equal(sapsystem.DiagnosticsAgent, system.Type)
	suite.NoError(err)
}

func (suite *SAPSystemTestSuite) TestDetectSystemId_Unknown() {
	ctx := context.TODO()
	appFS, err := mockDEVFileSystem()
	suite.NoError(err)
	mockCommand := new(mocks.MockCommandExecutor)
	mockWebServiceConnector := new(sapControlMocks.MockWebServiceConnector)

	mockWebServiceConnector.On("New", "01").Return(fakeNewWebService("HDB00", "UNKNOWN"))

	system, err := sapsystem.NewSAPSystem(ctx, appFS, mockCommand, mockWebServiceConnector, "/usr/sap/DEV")

	suite.Equal("-", system.ID)
	suite.Equal(sapsystem.Unknown, system.Type)
	suite.NoError(err)
}

func (suite *SAPSystemTestSuite) TestGetDatabases() {
	appFS := afero.NewMemMapFs()
	err := appFS.MkdirAll("/usr/sap/DEV/SYS/global/hdb/mdc/", 0755)
	suite.NoError(err)

	nameserverContent := []byte(`
# DATABASE:CONTAINER:USER:GROUP:USERID:GROUPID:HOST:SQLPORT:ACTIVE
PRD::::::hana01:30015:yes

DEV::::::hana01:30044:yes
ERR:::
`)

	err = afero.WriteFile(
		appFS, "/usr/sap/DEV/SYS/global/hdb/mdc/databases.lst",
		nameserverContent, 0644)
	suite.NoError(err)

	dbs, err := sapsystem.GetDatabases(appFS, "DEV")

	expectedDbs := []*sapsystem.DatabaseData{
		{
			Database:  "PRD",
			Container: "",
			User:      "",
			Group:     "",
			UserID:    "",
			GroupID:   "",
			Host:      "hana01",
			SQLPort:   "30015",
			Active:    "yes",
		},
		{
			Database:  "DEV",
			Container: "",
			User:      "",
			Group:     "",
			UserID:    "",
			GroupID:   "",
			Host:      "hana01",
			SQLPort:   "30044",
			Active:    "yes",
		},
	}

	suite.NoError(err)
	suite.Equal(len(dbs), 2)
	suite.ElementsMatch(expectedDbs, dbs)
}

func (suite *SAPSystemTestSuite) TestGetDBAddress() {
	s := &sapsystem.SAPSystem{Profile: sapsystem.SAPProfile{"SAPDBHOST": "localhost"}}
	addr, err := s.GetDBAddress()
	suite.NoError(err)
	suite.Equal("127.0.0.1", addr)
}

func (suite *SAPSystemTestSuite) TestGetDBAddress_ResolveError() {
	s := &sapsystem.SAPSystem{Profile: sapsystem.SAPProfile{"SAPDBHOST": "other"}}
	_, err := s.GetDBAddress()
	suite.EqualError(err, "could not resolve \"other\" hostname")
}

func (suite *SAPSystemTestSuite) TestNewSAPInstanceDatabase() {
	ctx := context.TODO()
	mockWebService := new(sapControlMocks.MockWebService)
	mockCommand := new(mocks.MockCommandExecutor)
	host, _ := os.Hostname()

	appFS := afero.NewMemMapFs()
	err := appFS.MkdirAll("/usr/sap/PRD/SYS/global/sapcontrol", 0755)
	suite.NoError(err)
	err = afero.WriteFile(
		appFS,
		"/usr/sap/PRD/SYS/global/sapcontrol/0.3_50013_50014_0_2_00_host1",
		[]byte("Host:otherhost Pid:100"),
		0644,
	)
	suite.NoError(err)
	err = afero.WriteFile(
		appFS,
		"/usr/sap/PRD/SYS/global/sapcontrol/0.3_50113_50114_0_3_01_host2",
		[]byte(fmt.Sprintf("Host:%s Pid:100", host)),
		0644,
	)
	suite.NoError(err)

	mockWebService.
		On("GetInstancePropertiesContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetInstancePropertiesResponse{
			Properties: []*sapcontrolapi.InstanceProperty{
				{
					Property:     "prop1",
					Propertytype: "type1",
					Value:        "value1",
				},
				{
					Property:     "SAPSYSTEMNAME",
					Propertytype: "string",
					Value:        "PRD",
				},
				{
					Property:     "SAPLOCALHOST",
					Propertytype: "string",
					Value:        "host2",
				},
				{
					Property:     "INSTANCE_NAME",
					Propertytype: "string",
					Value:        "HDB01",
				},
				{
					Property:     "SAPSYSTEM",
					Propertytype: "string",
					Value:        "01",
				},
			},
		}, nil).
		On("GetProcessListContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetProcessListResponse{
			Processes: []*sapcontrolapi.OSProcess{
				{
					Name:        "enserver",
					Description: "foobar",
					Dispstatus:  sapcontrolapi.STATECOLOR_GREEN,
					Textstatus:  "Running",
					Starttime:   "",
					Elapsedtime: "",
					Pid:         30787,
				},
				{
					Name:        "msg_server",
					Description: "foobar2",
					Dispstatus:  sapcontrolapi.STATECOLOR_YELLOW,
					Textstatus:  "Stopping",
					Starttime:   "",
					Elapsedtime: "",
					Pid:         30786,
				},
			},
		}, nil).
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetSystemInstanceListResponse{
			Instances: []*sapcontrolapi.SAPInstance{
				{
					Hostname:      "host1",
					InstanceNr:    0,
					HttpPort:      50013,
					HttpsPort:     50014,
					StartPriority: "0.3",
					Features:      "HDB|HDB_WORKER",
					Dispstatus:    sapcontrolapi.STATECOLOR_GREEN,
				},
				{
					Hostname:      "host2",
					InstanceNr:    1,
					HttpPort:      50113,
					HttpsPort:     50114,
					StartPriority: "0.3",
					Features:      "HDB|HDB_WORKER",
					Dispstatus:    sapcontrolapi.STATECOLOR_YELLOW,
				},
			},
		}, nil)

	mockCommand.On("Output", "/usr/bin/su", "-lc", "python /usr/sap/PRD/HDB01/exe/python_support/systemReplicationStatus.py --sapcontrol=1", "prdadm").Return(
		mockSystemReplicationStatus(), nil,
	)

	mockCommand.On("Output", "/usr/bin/su", "-lc", "python /usr/sap/PRD/HDB01/exe/python_support/landscapeHostConfiguration.py --sapcontrol=1", "prdadm").Return(
		mockLandscapeHostConfiguration(), nil,
	)

	mockCommand.On("Output", "/usr/bin/su", "-lc", "/usr/sap/PRD/HDB01/exe/hdbnsutil -sr_state -sapcontrol=1", "prdadm").Return(
		mockHdbnsutilSrstate(), nil,
	)

	sapInstance, _ := sapsystem.NewSAPInstance(ctx, mockWebService, mockCommand, appFS)

	expectedInstance := &sapsystem.SAPInstance{
		Name: "HDB01",
		Type: sapsystem.Database,
		Host: host,
		SAPControl: &sapsystem.SAPControl{
			Processes: []*sapcontrolapi.OSProcess{
				{
					Name:        "enserver",
					Description: "foobar",
					Dispstatus:  sapcontrolapi.STATECOLOR_GREEN,
					Textstatus:  "Running",
					Starttime:   "",
					Elapsedtime: "",
					Pid:         30787,
				},
				{
					Name:        "msg_server",
					Description: "foobar2",
					Dispstatus:  sapcontrolapi.STATECOLOR_YELLOW,
					Textstatus:  "Stopping",
					Starttime:   "",
					Elapsedtime: "",
					Pid:         30786,
				},
			},
			Properties: []*sapcontrolapi.InstanceProperty{
				{
					Property:     "prop1",
					Propertytype: "type1",
					Value:        "value1",
				},
				{
					Property:     "SAPSYSTEMNAME",
					Propertytype: "string",
					Value:        "PRD",
				},
				{
					Property:     "SAPLOCALHOST",
					Propertytype: "string",
					Value:        "host2",
				},
				{
					Property:     "INSTANCE_NAME",
					Propertytype: "string",
					Value:        "HDB01",
				},
				{
					Property:     "SAPSYSTEM",
					Propertytype: "string",
					Value:        "01",
				},
			},
			Instances: []*sapcontrolapi.SAPInstance{
				{
					Hostname:        "host1",
					InstanceNr:      0,
					HttpPort:        50013,
					HttpsPort:       50014,
					StartPriority:   "0.3",
					Features:        "HDB|HDB_WORKER",
					Dispstatus:      sapcontrolapi.STATECOLOR_GREEN,
					CurrentInstance: false,
				},
				{
					Hostname:        "host2",
					InstanceNr:      1,
					HttpPort:        50113,
					HttpsPort:       50114,
					StartPriority:   "0.3",
					Features:        "HDB|HDB_WORKER",
					Dispstatus:      sapcontrolapi.STATECOLOR_YELLOW,
					CurrentInstance: true,
				},
			},
		},
		SystemReplication: sapsystem.SystemReplication{
			"service/hana01/30001/SHIPPED_LOG_POSITION_TIME":             "2021-06-12 12:43:13.059197",
			"service/hana01/30001/LAST_LOG_POSITION_TIME":                "2021-06-12 12:43:13.059197",
			"service/hana01/30001/SHIPPED_FULL_REPLICA_DURATION":         "4060418",
			"service/hana01/30001/SHIPPED_LAST_DELTA_REPLICA_START_TIME": "-",
			"service/hana01/30001/SHIPPED_FULL_REPLICA_SIZE":             "1913069568",
			"service/hana01/30001/SITE_ID":                               "1",
			"service/hana01/30001/LAST_LOG_POSITION":                     "37624000",
			"service/hana01/30001/SECONDARY_ACTIVE_STATUS":               "YES",
			"service/hana01/30001/LAST_SAVEPOINT_LOG_POSITION":           "37624081",
			"service/hana01/30001/FULL_SYNC":                             "DISABLED",
			"service/hana01/30001/OPERATION_MODE":                        "logreplay",
			"service/hana01/30001/SHIPPED_LAST_FULL_REPLICA_START_TIME":  "2021-06-11 08:43:29.183940",
			"service/hana01/30001/LAST_SAVEPOINT_VERSION":                "510",
			"service/hana01/30001/LAST_SAVEPOINT_START_TIME":             "2021-06-12 12:45:11.401787",
			"service/hana01/30001/SERVICE_NAME":                          "nameserver",
			"service/hana01/30001/PORT":                                  "30001",
			"service/hana01/30001/SHIPPED_DELTA_REPLICA_COUNT":           "0",
			"service/hana01/30001/LAST_RESET_TIME":                       "2021-06-11 08:43:19.530050",
			"service/hana01/30001/SECONDARY_FAILOVER_COUNT":              "0",
			"service/hana01/30001/SHIPPED_FULL_REPLICA_COUNT":            "1",
			"service/hana01/30001/SHIPPED_LOG_BUFFERS_DURATION":          "139833248",
			"service/hana01/30001/REPLICATION_STATUS_DETAILS":            "",
			"service/hana01/30001/SHIPPED_DELTA_REPLICA_SIZE":            "0",
			"service/hana01/30001/SHIPPED_LOG_POSITION":                  "37624000",
			"service/hana01/30001/SHIPPED_DELTA_REPLICA_DURATION":        "0",
			"service/hana01/30001/RESET_COUNT":                           "0",
			"service/hana01/30001/SHIPPED_LAST_DELTA_REPLICA_SIZE":       "0",
			"service/hana01/30001/SHIPPED_LAST_DELTA_REPLICA_END_TIME":   "-",
			"service/hana01/30001/SITE_NAME":                             "Site1",
			"service/hana01/30001/SECONDARY_SITE_NAME":                   "Site2",
			"service/hana01/30001/REPLAYED_LOG_POSITION_TIME":            "2021-06-12 12:43:13.059197",
			"service/hana01/30001/SHIPPED_LAST_FULL_REPLICA_END_TIME":    "2021-06-11 08:43:33.244358",
			"service/hana01/30001/CREATION_TIME":                         "2021-06-11 08:43:19.530050",
			"site/2/SITE_NAME":                                           "Site2",
			"site/2/SOURCE_SITE_ID":                                      "1",
			"site/2/REPLICATION_MODE":                                    "SYNC",
			"site/2/REPLICATION_STATUS":                                  "ERROR",
			"overall_replication_status":                                 "ERROR",
			"site/1/REPLICATION_MODE":                                    "PRIMARY",
			"site/1/SITE_NAME":                                           "Site1",
			"local_site_id":                                              "1",
		},
		HostConfiguration: sapsystem.HostConfiguration{
			"hostActualRoles":        "worker",
			"removeStatus":           "",
			"nameServerConfigRole":   "master 1",
			"failoverStatus":         "",
			"hostConfigRoles":        "worker",
			"failoverActualGroup":    "default",
			"storageConfigPartition": "1",
			"host":                   "hana01",
			"indexServerConfigRole":  "worker",
			"failoverConfigGroup":    "default",
			"storageActualPartition": "1",
			"indexServerActualRole":  "master",
			"nameServerActualRole":   "master",
			"hostActive":             "yes",
			"workerActualGroups":     "default",
			"workerConfigGroups":     "default",
			"hostStatus":             "ok",
			"storagePartition":       "1",
		},
		HdbnsutilSRstate: sapsystem.HdbnsutilSRstate{
			"online":             "true",
			"mode":               "primary",
			"operation_mode":     "primary",
			"site_id":            "1",
			"site_name":          "Site1",
			"isSource":           "true",
			"isConsumer":         "false",
			"hasConsumers":       "true",
			"isTakeoverActive":   "false",
			"isPrimarySuspended": "false",
			"mapping/hana01": []interface{}{
				"Site2/hana02",
				"Site1/hana01",
			},
			"siteTier/Site1":            "1",
			"siteTier/Site2":            "2",
			"siteReplicationMode/Site1": "primary",
			"siteReplicationMode/Site2": "sync",
			"siteOperationMode/Site1":   "primary",
			"siteOperationMode/Site2":   "logreplay",
			"siteMapping/Site1":         "Site2",
		},
	}

	suite.Equal(expectedInstance, sapInstance)
}

func (suite *SAPSystemTestSuite) TestNewSAPInstanceApp() {
	ctx := context.TODO()
	mockWebService := new(sapControlMocks.MockWebService)

	host, _ := os.Hostname()

	appFS := afero.NewMemMapFs()
	err := appFS.MkdirAll("/usr/sap/PRD/SYS/global/sapcontrol", 0755)
	suite.NoError(err)
	err = afero.WriteFile(
		appFS,
		"/usr/sap/PRD/SYS/global/sapcontrol/0.3_50013_50014_0_2_00_host1",
		[]byte(fmt.Sprintf("Host:%s Pid:100", host)),
		0644,
	)
	suite.NoError(err)
	err = afero.WriteFile(
		appFS,
		"/usr/sap/PRD/SYS/global/sapcontrol/0.3_50113_50114_0_3_01_host2",
		[]byte("Host:otherhost Pid:100"),
		0644,
	)
	suite.NoError(err)

	mockWebService.
		On("GetInstancePropertiesContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetInstancePropertiesResponse{
			Properties: []*sapcontrolapi.InstanceProperty{
				{
					Property:     "prop1",
					Propertytype: "type1",
					Value:        "value1",
				},
				{
					Property:     "SAPSYSTEMNAME",
					Propertytype: "string",
					Value:        "PRD",
				},
				{
					Property:     "INSTANCE_NAME",
					Propertytype: "string",
					Value:        "HDB00",
				},
				{
					Property:     "SAPLOCALHOST",
					Propertytype: "string",
					Value:        "host1",
				},
				{
					Property:     "SAPSYSTEM",
					Propertytype: "string",
					Value:        "00",
				},
			},
		}, nil)

	mockWebService.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetProcessListResponse{
			Processes: []*sapcontrolapi.OSProcess{
				{
					Name:        "enserver",
					Description: "foobar",
					Dispstatus:  sapcontrolapi.STATECOLOR_GREEN,
					Textstatus:  "Running",
					Starttime:   "",
					Elapsedtime: "",
					Pid:         30787,
				},
				{
					Name:        "msg_server",
					Description: "foobar2",
					Dispstatus:  sapcontrolapi.STATECOLOR_YELLOW,
					Textstatus:  "Stopping",
					Starttime:   "",
					Elapsedtime: "",
					Pid:         30786,
				},
			},
		}, nil)

	mockWebService.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetSystemInstanceListResponse{
			Instances: []*sapcontrolapi.SAPInstance{
				{
					Hostname:      "host1",
					InstanceNr:    0,
					HttpPort:      50013,
					HttpsPort:     50014,
					StartPriority: "0.3",
					Features:      "MESSAGESERVER|ENQUE",
					Dispstatus:    sapcontrolapi.STATECOLOR_GREEN,
				},
				{
					Hostname:      "host2",
					InstanceNr:    1,
					HttpPort:      50113,
					HttpsPort:     50114,
					StartPriority: "0.3",
					Features:      "some other features",
					Dispstatus:    sapcontrolapi.STATECOLOR_YELLOW,
				},
			},
		}, nil)

	sapInstance, _ := sapsystem.NewSAPInstance(ctx, mockWebService, new(mocks.MockCommandExecutor), appFS)

	expectedInstance := &sapsystem.SAPInstance{
		Name: "HDB00",
		Type: sapsystem.Application,
		Host: host,
		SAPControl: &sapsystem.SAPControl{
			Processes: []*sapcontrolapi.OSProcess{
				{
					Name:        "enserver",
					Description: "foobar",
					Dispstatus:  sapcontrolapi.STATECOLOR_GREEN,
					Textstatus:  "Running",
					Starttime:   "",
					Elapsedtime: "",
					Pid:         30787,
				},
				{
					Name:        "msg_server",
					Description: "foobar2",
					Dispstatus:  sapcontrolapi.STATECOLOR_YELLOW,
					Textstatus:  "Stopping",
					Starttime:   "",
					Elapsedtime: "",
					Pid:         30786,
				},
			},
			Properties: []*sapcontrolapi.InstanceProperty{
				{
					Property:     "prop1",
					Propertytype: "type1",
					Value:        "value1",
				},
				{
					Property:     "SAPSYSTEMNAME",
					Propertytype: "string",
					Value:        "PRD",
				},
				{
					Property:     "INSTANCE_NAME",
					Propertytype: "string",
					Value:        "HDB00",
				},
				{
					Property:     "SAPLOCALHOST",
					Propertytype: "string",
					Value:        "host1",
				},
				{
					Property:     "SAPSYSTEM",
					Propertytype: "string",
					Value:        "00",
				},
			},
			Instances: []*sapcontrolapi.SAPInstance{
				{
					Hostname:        "host1",
					InstanceNr:      0,
					HttpPort:        50013,
					HttpsPort:       50014,
					StartPriority:   "0.3",
					Features:        "MESSAGESERVER|ENQUE",
					Dispstatus:      sapcontrolapi.STATECOLOR_GREEN,
					CurrentInstance: true,
				},
				{
					Hostname:        "host2",
					InstanceNr:      1,
					HttpPort:        50113,
					HttpsPort:       50114,
					StartPriority:   "0.3",
					Features:        "some other features",
					Dispstatus:      sapcontrolapi.STATECOLOR_YELLOW,
					CurrentInstance: false,
				},
			},
		},
		SystemReplication: sapsystem.SystemReplication(nil),
		HostConfiguration: sapsystem.HostConfiguration(nil),
		HdbnsutilSRstate:  sapsystem.HdbnsutilSRstate(nil),
	}

	suite.Equal(expectedInstance, sapInstance)
}

func (suite *SAPSystemTestSuite) TestGetSIDsString() {
	sysList := sapsystem.SAPSystemsList{
		&sapsystem.SAPSystem{
			SID: "PRD",
		},
	}

	suite.Equal("PRD", sysList.GetSIDsString())

	sysList = sapsystem.SAPSystemsList{
		&sapsystem.SAPSystem{
			SID: "PRD",
		},
		&sapsystem.SAPSystem{
			SID: "QAS",
		},
	}

	suite.Equal("PRD,QAS", sysList.GetSIDsString())
}

func (suite *SAPSystemTestSuite) TestFindSystemsNotFound() {
	appFS := afero.NewMemMapFs()
	// create test files and directories
	err := appFS.MkdirAll("/usr/sap/", 0755)
	suite.NoError(err)

	err = appFS.MkdirAll("/usr/sap/DEV1/", 0755)
	suite.NoError(err)

	systems, err := sapsystem.FindSystems(appFS)

	suite.Equal([]string{}, systems)
	suite.NoError(err)
}

func (suite *SAPSystemTestSuite) TestFindSystems() {
	appFS := afero.NewMemMapFs()
	// create test files and directories
	err := appFS.MkdirAll("/usr/sap/PRD/HDB00", 0755)
	suite.NoError(err)

	err = appFS.MkdirAll("/usr/sap/PRD/HDB01", 0755)
	suite.NoError(err)

	err = appFS.MkdirAll("/usr/sap/DEV/ASCS02", 0755)
	suite.NoError(err)

	err = appFS.MkdirAll("/usr/sap/DEV1/ASCS02", 0755)
	suite.NoError(err)

	err = appFS.MkdirAll("/usr/sap/DEV/SYS/BLA12", 0755)
	suite.NoError(err)

	err = appFS.MkdirAll("/usr/sap/DEV/PRD0", 0755)
	suite.NoError(err)

	systems, err := sapsystem.FindSystems(appFS)
	suite.ElementsMatch([]string{"/usr/sap/PRD", "/usr/sap/DEV"}, systems)
	suite.NoError(err)
}

func (suite *SAPSystemTestSuite) TestFindInstancesNotFound() {
	appFS := afero.NewMemMapFs()
	// create test files and directories
	err := appFS.MkdirAll("/usr/sap/DEV/SYS/BLA12", 0755)
	suite.NoError(err)

	instances, err := sapsystem.FindInstances(appFS, "/usr/sap/DEV")

	suite.Equal([][]string{}, instances)
	suite.NoError(err)
}

func (suite *SAPSystemTestSuite) TestFindInstances() {
	appFS := afero.NewMemMapFs()
	// create test files and directories
	err := appFS.MkdirAll("/usr/sap/DEV/ASCS02", 0755)
	suite.NoError(err)

	err = appFS.MkdirAll("/usr/sap/DEV/SYS/BLA12", 0755)
	suite.NoError(err)

	err = appFS.MkdirAll("/usr/sap/DEV/PRD0", 0755)
	suite.NoError(err)

	err = appFS.MkdirAll("/usr/sap/DEV/ERS10", 0755)
	suite.NoError(err)

	instances, err := sapsystem.FindInstances(appFS, "/usr/sap/DEV")
	expectedInstance := [][]string{
		{"ASCS02", "02"},
		{"ERS10", "10"},
	}
	suite.ElementsMatch(expectedInstance, instances)
	suite.NoError(err)
}

func (suite *SAPSystemTestSuite) TestFindProfilesNotFound() {
	appFS := afero.NewMemMapFs()
	// create test files and directories
	err := appFS.MkdirAll("/sapmnt/DEV/profile", 0755)
	suite.NoError(err)
	err = appFS.MkdirAll("/sapmnt/PRD/profile", 0755)
	suite.NoError(err)

	profiles, err := sapsystem.FindProfiles(appFS, "DEV")

	suite.Equal([]string{}, profiles)
	suite.NoError(err)
}

func (suite *SAPSystemTestSuite) TestFindProfiles() {
	appFS := afero.NewMemMapFs()
	// create test files and directories
	err := afero.WriteFile(appFS, "/sapmnt/DEV/profile/DEFAULT.1.PFL", []byte{}, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/DEV/profile/DEFAULT.PFL", []byte{}, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/DEV/profile/dev_profile", []byte{}, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/DEV/profile/dev_profile.1", []byte{}, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/DEV/profile/dev_profile.bak", []byte{}, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/PRD/profile/prd_profile", []byte{}, 0644)
	suite.NoError(err)

	profiles, err := sapsystem.FindProfiles(appFS, "DEV")
	expectedProfiles := []string{"DEFAULT.PFL", "dev_profile"}

	suite.ElementsMatch(expectedProfiles, profiles)
	suite.NoError(err)
}

func (suite *SAPSystemTestSuite) TestDetectType() {
	ctx := context.TODO()
	cases := []struct {
		instance     *sapcontrolapi.SAPInstance
		expectedType sapsystem.SystemType
	}{
		{
			instance: &sapcontrolapi.SAPInstance{
				Hostname: "host",
				Features: "MESSAGESERVER|ENQUE",
			},
			expectedType: sapsystem.Application,
		},
		{
			instance: &sapcontrolapi.SAPInstance{
				Hostname: "host",
				Features: "GATEWAY|MESSAGESERVER|ENQUE|WEBDISP",
			},
			expectedType: sapsystem.Application,
		},
		{
			instance: &sapcontrolapi.SAPInstance{
				Hostname: "host",
				Features: "ENQREP",
			},
			expectedType: sapsystem.Application,
		},
		{
			instance: &sapcontrolapi.SAPInstance{
				Hostname: "host",
				Features: "ABAP|GATEWAY|ICMAN|IGS",
			},
			expectedType: sapsystem.Application,
		},
		{
			instance: &sapcontrolapi.SAPInstance{
				Hostname: "host",
				Features: "J2EE|ICMAN|IGS",
			},
			expectedType: sapsystem.Application,
		},
		{
			instance: &sapcontrolapi.SAPInstance{
				Hostname: "host",
				Features: "J2EE|IGS",
			},
			expectedType: sapsystem.Application,
		},
		{
			instance: &sapcontrolapi.SAPInstance{
				Hostname: "host",
				Features: "SMDAGENT",
			},
			expectedType: sapsystem.DiagnosticsAgent,
		},
		{
			instance: &sapcontrolapi.SAPInstance{
				Hostname: "host",
				Features: "UNKNOWNFEATURE",
			},
			expectedType: sapsystem.Unknown,
		},
	}

	appFS := afero.NewMemMapFs()
	err := appFS.MkdirAll("/usr/sap/PRD/SYS/global/sapcontrol", 0755)
	suite.NoError(err)

	for _, tt := range cases {
		mockWebService := new(sapControlMocks.MockWebService)
		mockWebService.
			On("GetInstancePropertiesContext", ctx, mock.Anything).
			Return(&sapcontrolapi.GetInstancePropertiesResponse{
				Properties: []*sapcontrolapi.InstanceProperty{
					{
						Property:     "SAPSYSTEMNAME",
						Propertytype: "string",
						Value:        "PRD",
					},
					{
						Property:     "INSTANCE_NAME",
						Propertytype: "string",
						Value:        "ASCS00",
					},
					{
						Property:     "SAPLOCALHOST",
						Propertytype: "string",
						Value:        "host",
					},
					{
						Property:     "SAPSYSTEM",
						Propertytype: "string",
						Value:        "00",
					},
				},
			}, nil).
			On("GetProcessListContext", ctx, mock.Anything).
			Return(&sapcontrolapi.GetProcessListResponse{
				Processes: []*sapcontrolapi.OSProcess{},
			}, nil).
			On("GetSystemInstanceListContext", ctx, mock.Anything).
			Return(&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{tt.instance},
			}, nil)

		mockCommand := new(mocks.MockCommandExecutor)
		instance, err := sapsystem.NewSAPInstance(ctx, mockWebService, mockCommand, appFS)

		suite.NoError(err)
		suite.Equal(tt.expectedType, instance.Type)
	}
}

func (suite *SAPSystemTestSuite) TestDetectType_Database() {
	ctx := context.TODO()

	appFS := afero.NewMemMapFs()
	err := appFS.MkdirAll("/usr/sap/HDB/SYS/global/sapcontrol", 0755)
	suite.NoError(err)

	mockWebService := new(sapControlMocks.MockWebService)
	mockWebService.
		On("GetInstancePropertiesContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetInstancePropertiesResponse{
			Properties: []*sapcontrolapi.InstanceProperty{
				{
					Property:     "SAPSYSTEMNAME",
					Propertytype: "string",
					Value:        "HDB",
				},
				{
					Property:     "INSTANCE_NAME",
					Propertytype: "string",
					Value:        "HDB00",
				},
				{
					Property:     "SAPLOCALHOST",
					Propertytype: "string",
					Value:        "host2",
				},
				{
					Property:     "SAPSYSTEM",
					Propertytype: "string",
					Value:        "00",
				},
			},
		}, nil).
		On("GetProcessListContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetProcessListResponse{
			Processes: []*sapcontrolapi.OSProcess{},
		}, nil).
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetSystemInstanceListResponse{
			Instances: []*sapcontrolapi.SAPInstance{
				{
					Hostname: "host1",
					Features: "other",
				},
				{
					Hostname: "host2",
					Features: "HDB|HDB_WORKER",
				},
			},
		}, nil)

	mockCommand := new(mocks.MockCommandExecutor)
	mockCommand.
		On("Output", "/usr/bin/su", "-lc", "python /usr/sap/HDB/HDB00/exe/python_support/systemReplicationStatus.py --sapcontrol=1", "hdbadm").
		Return(mockSystemReplicationStatus(), nil).
		On("Output", "/usr/bin/su", "-lc", "python /usr/sap/HDB/HDB00/exe/python_support/landscapeHostConfiguration.py --sapcontrol=1", "hdbadm").
		Return(mockLandscapeHostConfiguration(), nil).
		On("Output", "/usr/bin/su", "-lc", "/usr/sap/HDB/HDB00/exe/hdbnsutil -sr_state -sapcontrol=1", "hdbadm").
		Return(mockHdbnsutilSrstate(), nil)

	instance, err := sapsystem.NewSAPInstance(ctx, mockWebService, mockCommand, appFS)

	suite.NoError(err)
	suite.Equal(sapsystem.Database, instance.Type)
}

func (suite *SAPSystemTestSuite) TestSapControl_NoSapcontrolFolder() {
	ctx := context.TODO()
	appFS := afero.NewMemMapFs()
	mockWebService := fakeNewWebService("ASCS01", "")

	_, err := sapsystem.NewSAPControl(ctx, mockWebService, appFS, "")
	suite.EqualError(
		err,
		"Error finding current instance: sapcontrol folder not found: "+
			"open /usr/sap/DEV/SYS/global/sapcontrol: file does not exist")
}

func (suite *SAPSystemTestSuite) TestSapControl_MissingProperties() {
	ctx := context.TODO()
	appFS := afero.NewMemMapFs()
	cases := []struct {
		properties          []*sapcontrolapi.InstanceProperty
		expectedMissingProp string
	}{
		{
			properties:          []*sapcontrolapi.InstanceProperty{},
			expectedMissingProp: "SAPSYSTEMNAME",
		},
		{
			properties: []*sapcontrolapi.InstanceProperty{
				{
					Property:     "SAPSYSTEMNAME",
					Propertytype: "string",
					Value:        "DEV",
				},
			},
			expectedMissingProp: "SAPSYSTEM",
		},
		{
			properties: []*sapcontrolapi.InstanceProperty{
				{
					Property:     "SAPSYSTEMNAME",
					Propertytype: "string",
					Value:        "DEV",
				},
				{
					Property:     "SAPSYSTEM",
					Propertytype: "string",
					Value:        "00",
				},
			},
			expectedMissingProp: "SAPLOCALHOST",
		},
	}

	for _, tt := range cases {
		mockWebService := new(sapControlMocks.MockWebService)
		mockWebService.
			On("GetInstancePropertiesContext", ctx, mock.Anything).
			Return(&sapcontrolapi.GetInstancePropertiesResponse{
				Properties: tt.properties,
			}, nil).
			On("GetProcessListContext", ctx, mock.Anything).
			Return(&sapcontrolapi.GetProcessListResponse{
				Processes: []*sapcontrolapi.OSProcess{},
			}, nil).
			On("GetSystemInstanceListContext", ctx, mock.Anything).
			Return(&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{},
			}, nil)

		_, err := sapsystem.NewSAPControl(ctx, mockWebService, appFS, "")
		suite.EqualError(
			err,
			fmt.Sprintf("Error finding current instance: Property %s not found", tt.expectedMissingProp))
	}
}
