//nolint:exhaustruct,gosec,nosnakecase,gochecknoglobals,lll,dupl
package sapsystem

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	sapSystemMocks "github.com/trento-project/agent/internal/sapsystem/mocks"
	"github.com/trento-project/agent/internal/sapsystem/sapcontrol"
	sapControlMocks "github.com/trento-project/agent/internal/sapsystem/sapcontrol/mocks"
)

type SAPSystemTestSuite struct {
	suite.Suite
	attempts int
}

func TestSAPSystemTestSuite(t *testing.T) {
	testSuite := new(SAPSystemTestSuite)
	suite.Run(t, testSuite)
}

func (suite *SAPSystemTestSuite) SetupSuite() {
	suite.attempts = 0
}

func (suite *SAPSystemTestSuite) fakeNewWebService(instNumber string) sapcontrol.WebService {
	var instance string

	mockWebService := new(sapControlMocks.WebService)

	defer func() {
		suite.attempts++
	}()

	if suite.attempts == 0 {
		instance = "ASCS01"
	} else if suite.attempts == 1 {
		instance = "ERS02"
	}

	mockWebService.On("GetInstanceProperties").Return(&sapcontrol.GetInstancePropertiesResponse{
		Properties: []*sapcontrol.InstanceProperty{
			{
				Property:     "SAPSYSTEMNAME",
				Propertytype: "string",
				Value:        "DEV",
			},
			{
				Property:     "INSTANCE_NAME",
				Propertytype: "string",
				Value:        instance,
			},
			{
				Property:     "SAPLOCALHOST",
				Propertytype: "string",
				Value:        "host",
			},
		},
	}, nil)

	mockWebService.On("GetProcessList").Return(&sapcontrol.GetProcessListResponse{
		Processes: []*sapcontrol.OSProcess{},
	}, nil)

	mockWebService.On("GetSystemInstanceList").Return(&sapcontrol.GetSystemInstanceListResponse{
		Instances: []*sapcontrol.SAPInstance{},
	}, nil)

	return mockWebService
}

func (suite *SAPSystemTestSuite) TestNewSAPSystemd() {

	mockCommand := new(sapSystemMocks.CustomCommand)

	customExecCommand = mockCommand.Execute
	newWebService = suite.fakeNewWebService

	appFS := afero.NewMemMapFs()
	err := appFS.MkdirAll("/usr/sap/DEV/ASCS01", 0755)
	suite.NoError(err)
	err = appFS.MkdirAll("/usr/sap/DEV/ERS02", 0755)
	suite.NoError(err)

	profileFile, _ := os.Open("../../test/sap_profile_default")
	profileContent, _ := io.ReadAll(profileFile)

	err = appFS.MkdirAll("/usr/sap/DEV/SYS/profile", 0755)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/usr/sap/DEV/SYS/profile/DEFAULT.PFL", profileContent, 0644)
	suite.NoError(err)

	expectedProfile := SAPProfile{
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

	cmd := fmt.Sprintf(sappfparCmd, "DEV")
	mockCommand.On("Execute", "su", "-lc", cmd, "devadm").Return(mockSappfpar())

	system, err := NewSAPSystem(appFS, "/usr/sap/DEV")

	suite.Equal(Unknown, system.Type)
	suite.Contains(system.Instances[0].Name, "ASCS01")
	suite.Contains(system.Instances[1].Name, "ERS02")
	suite.Equal(system.Profile, expectedProfile)
	suite.NoError(err)
}

func mockSystemReplicationStatus() *exec.Cmd {
	sFile, _ := os.Open("../../test/system_replication_status")
	content, _ := io.ReadAll(sFile)
	return exec.Command("echo", string(content))
}

func mockLandscapeHostConfiguration() *exec.Cmd {
	lFile, _ := os.Open("../../test/landscape_host_configuration")
	content, _ := io.ReadAll(lFile)
	return exec.Command("echo", string(content))
}

func mockHdbnsutilSrstate() *exec.Cmd {
	lFile, _ := os.Open("../../test/hdbnsutil_srstate")
	content, _ := io.ReadAll(lFile)
	return exec.Command("echo", string(content))
}

func mockSappfpar() *exec.Cmd {
	return exec.Command("echo", "-n", "systemId")
}

func (suite *SAPSystemTestSuite) TestSetSystemIdDatabased() {
	appFS := afero.NewMemMapFs()
	err := appFS.MkdirAll("/usr/sap/DEV/SYS/global/hdb/custom/config/", 0755)
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

	system := &SAPSystem{
		Type: Database,
		SID:  "DEV",
	}

	system, err = setSystemID(appFS, system)

	suite.NoError(err)
	suite.Equal("089d1a278481b86e821237f8e98e6de7", system.ID)
}

func (suite *SAPSystemTestSuite) TestSetSystemIdApplicationd() {
	appFS := afero.NewMemMapFs()
	mockCommand := new(sapSystemMocks.CustomCommand)

	customExecCommand = mockCommand.Execute
	cmd := fmt.Sprintf(sappfparCmd, "DEV")
	mockCommand.On("Execute", "su", "-lc", cmd, "devadm").Return(mockSappfpar())

	system := &SAPSystem{
		Type: Application,
		SID:  "DEV",
	}

	system, err := setSystemID(appFS, system)

	suite.NoError(err)
	suite.Equal("089d1a278481b86e821237f8e98e6de7", system.ID)
}

func (suite *SAPSystemTestSuite) TestSetSystemIdOtherd() {
	appFS := afero.NewMemMapFs()

	system := &SAPSystem{
		Type: Unknown,
		SID:  "DEV",
	}

	system, err := setSystemID(appFS, system)

	suite.NoError(err)
	suite.Equal("-", system.ID)
}

func (suite *SAPSystemTestSuite) TestSetSystemIdDiagnosticsd() {
	appFS := afero.NewMemMapFs()
	err := appFS.MkdirAll("/etc", 0755)
	suite.NoError(err)

	machineIDContent := []byte(`dummy-machine-id`)

	err = afero.WriteFile(
		appFS, "/etc/machine-id",
		machineIDContent, 0644)
	suite.NoError(err)

	system := &SAPSystem{
		Type: DiagnosticsAgent,
		SID:  "DAA",
	}

	system, err = setSystemID(appFS, system)

	suite.NoError(err)
	suite.Equal("d3d5dd5ec501127e0011a2531e3b11ff", system.ID)
}

func (suite *SAPSystemTestSuite) TestGetDatabasesd() {
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

	dbs, err := getDatabases(appFS, "DEV")

	expectedDbs := []*DatabaseData{
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

func (suite *SAPSystemTestSuite) TestGetDBAddressd() {
	s := &SAPSystem{Profile: SAPProfile{"SAPDBHOST": "localhost"}}
	addr, err := getDBAddress(s)
	suite.NoError(err)
	suite.Equal("127.0.0.1", addr)
}

func (suite *SAPSystemTestSuite) TestGetDBAddress_ResolveErrord() {
	s := &SAPSystem{Profile: SAPProfile{"SAPDBHOST": "other"}}
	_, err := getDBAddress(s)
	suite.EqualError(err, "could not resolve \"other\" hostname")
}

func (suite *SAPSystemTestSuite) TestNewSAPInstanceDatabased() {
	mockWebService := new(sapControlMocks.WebService)
	mockCommand := new(sapSystemMocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockWebService.On("GetInstanceProperties").Return(&sapcontrol.GetInstancePropertiesResponse{
		Properties: []*sapcontrol.InstanceProperty{
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
				Value:        "host1",
			},
			{
				Property:     "INSTANCE_NAME",
				Propertytype: "string",
				Value:        "HDB00",
			},
		},
	}, nil)

	mockWebService.On("GetProcessList").Return(&sapcontrol.GetProcessListResponse{
		Processes: []*sapcontrol.OSProcess{
			{
				Name:        "enserver",
				Description: "foobar",
				Dispstatus:  sapcontrol.STATECOLOR_GREEN,
				Textstatus:  "Running",
				Starttime:   "",
				Elapsedtime: "",
				Pid:         30787,
			},
			{
				Name:        "msg_server",
				Description: "foobar2",
				Dispstatus:  sapcontrol.STATECOLOR_YELLOW,
				Textstatus:  "Stopping",
				Starttime:   "",
				Elapsedtime: "",
				Pid:         30786,
			},
		},
	}, nil)

	mockWebService.On("GetSystemInstanceList").Return(&sapcontrol.GetSystemInstanceListResponse{
		Instances: []*sapcontrol.SAPInstance{
			{
				Hostname:      "host1",
				InstanceNr:    0,
				HttpPort:      50013,
				HttpsPort:     50014,
				StartPriority: "0.3",
				Features:      "HDB|HDB_WORKER",
				Dispstatus:    sapcontrol.STATECOLOR_GREEN,
			},
			{
				Hostname:      "host2",
				InstanceNr:    1,
				HttpPort:      50113,
				HttpsPort:     50114,
				StartPriority: "0.3",
				Features:      "HDB|HDB_WORKER",
				Dispstatus:    sapcontrol.STATECOLOR_YELLOW,
			},
		},
	}, nil)

	mockCommand.On("Execute", "su", "-lc", "python /usr/sap/PRD/HDB00/exe/python_support/systemReplicationStatus.py --sapcontrol=1", "prdadm").Return(
		mockSystemReplicationStatus(),
	)

	mockCommand.On("Execute", "su", "-lc", "python /usr/sap/PRD/HDB00/exe/python_support/landscapeHostConfiguration.py --sapcontrol=1", "prdadm").Return(
		mockLandscapeHostConfiguration(),
	)

	mockCommand.On("Execute", "su", "-lc", "/usr/sap/PRD/HDB00/exe/hdbnsutil -sr_state -sapcontrol=1", "prdadm").Return(
		mockHdbnsutilSrstate(),
	)

	sapInstance, _ := NewSAPInstance(mockWebService)
	host, _ := os.Hostname()

	expectedInstance := &SAPInstance{
		Name: "HDB00",
		Type: Database,
		Host: host,
		SAPControl: &SAPControl{
			webService: mockWebService,
			Processes: []*sapcontrol.OSProcess{
				{
					Name:        "enserver",
					Description: "foobar",
					Dispstatus:  sapcontrol.STATECOLOR_GREEN,
					Textstatus:  "Running",
					Starttime:   "",
					Elapsedtime: "",
					Pid:         30787,
				},
				{
					Name:        "msg_server",
					Description: "foobar2",
					Dispstatus:  sapcontrol.STATECOLOR_YELLOW,
					Textstatus:  "Stopping",
					Starttime:   "",
					Elapsedtime: "",
					Pid:         30786,
				},
			},
			Properties: []*sapcontrol.InstanceProperty{
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
					Value:        "host1",
				},
				{
					Property:     "INSTANCE_NAME",
					Propertytype: "string",
					Value:        "HDB00",
				},
			},
			Instances: []*sapcontrol.SAPInstance{
				{
					Hostname:      "host1",
					InstanceNr:    0,
					HttpPort:      50013,
					HttpsPort:     50014,
					StartPriority: "0.3",
					Features:      "HDB|HDB_WORKER",
					Dispstatus:    sapcontrol.STATECOLOR_GREEN,
				},
				{
					Hostname:      "host2",
					InstanceNr:    1,
					HttpPort:      50113,
					HttpsPort:     50114,
					StartPriority: "0.3",
					Features:      "HDB|HDB_WORKER",
					Dispstatus:    sapcontrol.STATECOLOR_YELLOW,
				},
			},
		},
		SystemReplication: SystemReplication{
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
		HostConfiguration: HostConfiguration{
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
		HdbnsutilSRstate: HdbnsutilSRstate{
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

func (suite *SAPSystemTestSuite) TestNewSAPInstanceAppd() {
	mockWebService := new(sapControlMocks.WebService)

	mockWebService.On("GetInstanceProperties").Return(&sapcontrol.GetInstancePropertiesResponse{
		Properties: []*sapcontrol.InstanceProperty{
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
		},
	}, nil)

	mockWebService.On("GetProcessList").Return(&sapcontrol.GetProcessListResponse{
		Processes: []*sapcontrol.OSProcess{
			{
				Name:        "enserver",
				Description: "foobar",
				Dispstatus:  sapcontrol.STATECOLOR_GREEN,
				Textstatus:  "Running",
				Starttime:   "",
				Elapsedtime: "",
				Pid:         30787,
			},
			{
				Name:        "msg_server",
				Description: "foobar2",
				Dispstatus:  sapcontrol.STATECOLOR_YELLOW,
				Textstatus:  "Stopping",
				Starttime:   "",
				Elapsedtime: "",
				Pid:         30786,
			},
		},
	}, nil)

	mockWebService.On("GetSystemInstanceList").Return(&sapcontrol.GetSystemInstanceListResponse{
		Instances: []*sapcontrol.SAPInstance{
			{
				Hostname:      "host1",
				InstanceNr:    0,
				HttpPort:      50013,
				HttpsPort:     50014,
				StartPriority: "0.3",
				Features:      "MESSAGESERVER|ENQUE",
				Dispstatus:    sapcontrol.STATECOLOR_GREEN,
			},
			{
				Hostname:      "host2",
				InstanceNr:    1,
				HttpPort:      50113,
				HttpsPort:     50114,
				StartPriority: "0.3",
				Features:      "some other features",
				Dispstatus:    sapcontrol.STATECOLOR_YELLOW,
			},
		},
	}, nil)

	sapInstance, _ := NewSAPInstance(mockWebService)
	host, _ := os.Hostname()

	expectedInstance := &SAPInstance{
		Name: "HDB00",
		Type: Application,
		Host: host,
		SAPControl: &SAPControl{
			webService: mockWebService,
			Processes: []*sapcontrol.OSProcess{
				{
					Name:        "enserver",
					Description: "foobar",
					Dispstatus:  sapcontrol.STATECOLOR_GREEN,
					Textstatus:  "Running",
					Starttime:   "",
					Elapsedtime: "",
					Pid:         30787,
				},
				{
					Name:        "msg_server",
					Description: "foobar2",
					Dispstatus:  sapcontrol.STATECOLOR_YELLOW,
					Textstatus:  "Stopping",
					Starttime:   "",
					Elapsedtime: "",
					Pid:         30786,
				},
			},
			Properties: []*sapcontrol.InstanceProperty{
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
			},
			Instances: []*sapcontrol.SAPInstance{
				{
					Hostname:      "host1",
					InstanceNr:    0,
					HttpPort:      50013,
					HttpsPort:     50014,
					StartPriority: "0.3",
					Features:      "MESSAGESERVER|ENQUE",
					Dispstatus:    sapcontrol.STATECOLOR_GREEN,
				},
				{
					Hostname:      "host2",
					InstanceNr:    1,
					HttpPort:      50113,
					HttpsPort:     50114,
					StartPriority: "0.3",
					Features:      "some other features",
					Dispstatus:    sapcontrol.STATECOLOR_YELLOW,
				},
			},
		},
		SystemReplication: SystemReplication(nil),
		HostConfiguration: HostConfiguration(nil),
		HdbnsutilSRstate:  HdbnsutilSRstate(nil),
	}

	suite.Equal(expectedInstance, sapInstance)
}

func (suite *SAPSystemTestSuite) TestGetSIDsStringd() {
	sysList := SAPSystemsList{
		&SAPSystem{
			SID: "PRD",
		},
	}

	suite.Equal("PRD", sysList.GetSIDsString())

	sysList = SAPSystemsList{
		&SAPSystem{
			SID: "PRD",
		},
		&SAPSystem{
			SID: "QAS",
		},
	}

	suite.Equal("PRD,QAS", sysList.GetSIDsString())
}

func (suite *SAPSystemTestSuite) TestFindSystemsNotFoundd() {
	appFS := afero.NewMemMapFs()
	// create test files and directories
	err := appFS.MkdirAll("/usr/sap/", 0755)
	suite.NoError(err)

	err = appFS.MkdirAll("/usr/sap/DEV1/", 0755)
	suite.NoError(err)

	systems, _ := findSystems(appFS)

	suite.Equal([]string{}, systems)
}

func (suite *SAPSystemTestSuite) TestFindSystemsd() {
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

	systems, _ := findSystems(appFS)
	suite.ElementsMatch([]string{"/usr/sap/PRD", "/usr/sap/DEV"}, systems)
}

func (suite *SAPSystemTestSuite) TestFindInstancesNotFoundd() {
	appFS := afero.NewMemMapFs()
	// create test files and directories
	err := appFS.MkdirAll("/usr/sap/DEV/SYS/BLA12", 0755)
	suite.NoError(err)

	instances, _ := findInstances(appFS, "/usr/sap/DEV")

	suite.Equal([][]string{}, instances)
}

func (suite *SAPSystemTestSuite) TestFindInstancesd() {
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

	instances, _ := findInstances(appFS, "/usr/sap/DEV")
	expectedInstance := [][]string{
		{"ASCS02", "02"},
		{"ERS10", "10"},
	}
	suite.ElementsMatch(expectedInstance, instances)
}

func (suite *SAPSystemTestSuite) TestDetectType_Databased() {
	sapControl := &SAPControl{
		Properties: []*sapcontrol.InstanceProperty{
			{
				Property:     "SAPLOCALHOST",
				Propertytype: "string",
				Value:        "host2",
			},
		},
		Instances: []*sapcontrol.SAPInstance{
			{
				Hostname: "host1",
				Features: "other",
			},
			{
				Hostname: "host2",
				Features: "HDB|HDB_WORKER",
			},
		},
	}

	instanceType, err := detectType(sapControl)

	suite.NoError(err)
	suite.Equal(Database, instanceType)
}

func (suite *SAPSystemTestSuite) TestDetectType_Applicationd() {
	sapControl := &SAPControl{
		Properties: []*sapcontrol.InstanceProperty{
			{
				Property:     "SAPLOCALHOST",
				Propertytype: "string",
				Value:        "host1",
			},
		},
		Instances: []*sapcontrol.SAPInstance{
			{
				Hostname: "host1",
				Features: "MESSAGESERVER|ENQUE",
			},
		},
	}

	instanceType, err := detectType(sapControl)

	suite.NoError(err)
	suite.Equal(Application, instanceType)

	sapControl.Instances = []*sapcontrol.SAPInstance{
		{
			Hostname: "host1",
			Features: "ENQREP",
		},
	}

	instanceType, err = detectType(sapControl)

	suite.NoError(err)
	suite.Equal(Application, instanceType)

	sapControl.Instances = []*sapcontrol.SAPInstance{
		{
			Hostname: "host1",
			Features: "ABAP|GATEWAY|ICMAN|IGS",
		},
	}

	instanceType, err = detectType(sapControl)

	suite.NoError(err)
	suite.Equal(Application, instanceType)
}

func (suite *SAPSystemTestSuite) TestDetectType_Diagnosticsd() {
	sapControl := &SAPControl{
		Properties: []*sapcontrol.InstanceProperty{
			{
				Property:     "SAPLOCALHOST",
				Propertytype: "string",
				Value:        "host1",
			},
		},
		Instances: []*sapcontrol.SAPInstance{
			{
				Hostname: "host1",
				Features: "SMDAGENT",
			},
		},
	}

	instanceType, err := detectType(sapControl)

	suite.NoError(err)
	suite.Equal(DiagnosticsAgent, instanceType)
}

func (suite *SAPSystemTestSuite) TestDetectType_Unknownd() {
	sapControl := &SAPControl{
		Properties: []*sapcontrol.InstanceProperty{
			{
				Property:     "SAPLOCALHOST",
				Propertytype: "string",
				Value:        "host2",
			},
		},
		Instances: []*sapcontrol.SAPInstance{
			{
				Hostname: "host1",
				Features: "other",
			},
			{
				Hostname: "host2",
				Features: "another",
			},
		},
	}

	instanceType, err := detectType(sapControl)

	suite.NoError(err)
	suite.Equal(Unknown, instanceType)
}
