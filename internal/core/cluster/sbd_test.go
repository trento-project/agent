package cluster_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cluster"
	"github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type SbdTestSuite struct {
	suite.Suite
}

func TestSbdTestSuite(t *testing.T) {
	suite.Run(t, new(SbdTestSuite))
}

func mockSbdDump() []byte {
	output := `==Dumping header on disk /dev/vdc
Header version     : 2.1
UUID               : 541bdcea-16af-44a4-8ab9-6a98602e65ca
Number of slots    : 255
Sector size        : 512
Timeout (watchdog) : 5
Timeout (allocate) : 2
Timeout (loop)     : 1
Timeout (msgwait)  : 10
==Header on disk /dev/vdc is dumped`
	return []byte(output)
}

func mockSbdDumpErr() []byte {
	output := `==Dumping header on disk /dev/vdc
Header version     : 2.1
UUID               : 541bdcea-16af-44a4-8ab9-6a98602e65ca
==Number of slots on disk /dev/vdb NOT dumped
sbd failed; please check the logs.`

	return []byte(output)
}

func mockSbdList() []byte {
	output := `0	hana01	clear
1	hana02	clear`
	return []byte(output)
}

func mockSbdListErr() []byte {
	output := `== disk /dev/vdxx unreadable!
sbd failed; please check the logs.`

	return []byte(output)
}

func (suite *SbdTestSuite) TestLoadDeviceData() {
	mockCommand := new(mocks.CommandExecutor)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdList(), nil)

	s := cluster.NewSBDDevice(mockCommand, "/bin/sbd", "/dev/vdc")

	err := s.LoadDeviceData()

	expectedDevice := cluster.SBDDevice{
		Device: "/dev/vdc",
		Status: "healthy",
		Dump: cluster.SBDDump{
			Header:          "2.1",
			UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
			Slots:           255,
			SectorSize:      512,
			TimeoutWatchdog: 5,
			TimeoutAllocate: 2,
			TimeoutLoop:     1,
			TimeoutMsgwait:  10,
		},
		List: []*cluster.SBDNode{
			{
				ID:     0,
				Name:   "hana01",
				Status: "clear",
			},
			{
				ID:     1,
				Name:   "hana02",
				Status: "clear",
			},
		},
	}

	suite.EqualExportedValues(expectedDevice, s)
	suite.NoError(err)
}

func (suite *SbdTestSuite) TestLoadDeviceDataDumpError() {
	mockCommand := new(mocks.CommandExecutor)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDumpErr(), errors.New("error"))
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdList(), nil)

	s := cluster.NewSBDDevice(mockCommand, "/bin/sbd", "/dev/vdc")

	err := s.LoadDeviceData()

	expectedDevice := cluster.SBDDevice{
		Device: "/dev/vdc",
		Status: "unhealthy",
		Dump: cluster.SBDDump{
			Header:          "2.1",
			UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
			Slots:           0,
			SectorSize:      0,
			TimeoutWatchdog: 0,
			TimeoutAllocate: 0,
			TimeoutLoop:     0,
			TimeoutMsgwait:  0,
		},
		List: []*cluster.SBDNode{
			{
				ID:     0,
				Name:   "hana01",
				Status: "clear",
			},
			{
				ID:     1,
				Name:   "hana02",
				Status: "clear",
			},
		},
	}

	suite.EqualExportedValues(expectedDevice, s)
	suite.EqualError(err, "sbd dump command error: error")
}

func (suite *SbdTestSuite) TestLoadDeviceDataListError() {
	mockCommand := new(mocks.CommandExecutor)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdListErr(), errors.New("error"))

	s := cluster.NewSBDDevice(mockCommand, "/bin/sbd", "/dev/vdc")

	err := s.LoadDeviceData()

	expectedDevice := cluster.SBDDevice{
		Device: "/dev/vdc",
		Status: "healthy",
		Dump: cluster.SBDDump{
			Header:          "2.1",
			UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
			Slots:           255,
			SectorSize:      512,
			TimeoutWatchdog: 5,
			TimeoutAllocate: 2,
			TimeoutLoop:     1,
			TimeoutMsgwait:  10,
		},
		List: []*cluster.SBDNode{},
	}

	suite.EqualExportedValues(expectedDevice, s)
	suite.EqualError(err, "sbd list command error: error")
}

func (suite *SbdTestSuite) TestLoadDeviceDataError() {
	mockCommand := new(mocks.CommandExecutor)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDumpErr(), errors.New("error"))
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdListErr(), errors.New("error"))

	s := cluster.NewSBDDevice(mockCommand, "/bin/sbd", "/dev/vdc")

	err := s.LoadDeviceData()

	expectedDevice := cluster.SBDDevice{
		Device: "/dev/vdc",
		Status: "unhealthy",
		Dump: cluster.SBDDump{
			Header:          "2.1",
			UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
			Slots:           0,
			SectorSize:      0,
			TimeoutWatchdog: 0,
			TimeoutAllocate: 0,
			TimeoutLoop:     0,
			TimeoutMsgwait:  0,
		},
		List: []*cluster.SBDNode{},
	}

	suite.EqualExportedValues(expectedDevice, s)
	suite.EqualError(err, "sbd dump command error: error;sbd list command error: error")
}

func (suite *SbdTestSuite) TestLoadSbdConfig() {
	sbdConfigVariants := []string{"sbd_config", "sbd_config_quoted_devices"}

	for _, sbdConfigVariant := range sbdConfigVariants {
		sbdConfig, err := cluster.LoadSbdConfig(helpers.GetFixturePath(fmt.Sprintf("discovery/cluster/sbd/%s", sbdConfigVariant)))

		expectedConfig := map[string]string{
			"SBD_OPTS":                "",
			"SBD_PACEMAKER":           "yes",
			"SBD_STARTMODE":           "always",
			"SBD_DELAY_START":         "no",
			"SBD_WATCHDOG_DEV":        "/dev/watchdog",
			"SBD_WATCHDOG_TIMEOUT":    "5",
			"SBD_TIMEOUT_ACTION":      "flush,reboot",
			"SBD_MOVE_TO_ROOT_CGROUP": "auto",
			"SBD_DEVICE":              "/dev/vdc;/dev/vdb",
			"AN_INTEGER":              "42",
			"TEST":                    "Value",
			"TEST2":                   "Value2",
		}

		suite.Equal(expectedConfig, sbdConfig)
		suite.NoError(err)
	}
}

func (suite *SbdTestSuite) TestLoadSbdConfigError() {
	sbdConfig, err := cluster.LoadSbdConfig("notexist")

	expectedConfig := map[string]string(nil)

	suite.Equal(expectedConfig, sbdConfig)
	suite.EqualError(err, "could not open sbd config file: open notexist: no such file or directory")
}

func (suite *SbdTestSuite) TestLoadSbdConfigParsingError() {
	sbdConfig, err := cluster.LoadSbdConfig(helpers.GetFixturePath("discovery/cluster/sbd/sbd_config_invalid"))

	expectedConfig := map[string]string(nil)

	suite.Equal(expectedConfig, sbdConfig)
	suite.EqualError(err, "could not parse sbd config file: error on line 1: missing =")
}

func (suite *SbdTestSuite) TestNewSBD() {
	mockCommand := new(mocks.CommandExecutor)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdList(), nil)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdb", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdb", "list").Return(mockSbdList(), nil)

	s, err := cluster.NewSBD(mockCommand, "/bin/sbd", helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"))

	expectedSbd := cluster.SBD{
		Config: map[string]string{
			"SBD_OPTS":                "",
			"SBD_PACEMAKER":           "yes",
			"SBD_STARTMODE":           "always",
			"SBD_DELAY_START":         "no",
			"SBD_WATCHDOG_DEV":        "/dev/watchdog",
			"SBD_WATCHDOG_TIMEOUT":    "5",
			"SBD_TIMEOUT_ACTION":      "flush,reboot",
			"SBD_MOVE_TO_ROOT_CGROUP": "auto",
			"SBD_DEVICE":              "/dev/vdc;/dev/vdb",
			"AN_INTEGER":              "42",
			"TEST":                    "Value",
			"TEST2":                   "Value2",
		},
		Devices: []*cluster.SBDDevice{
			{
				Device: "/dev/vdc",
				Status: "healthy",
				Dump: cluster.SBDDump{
					Header:          "2.1",
					UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
					Slots:           255,
					SectorSize:      512,
					TimeoutWatchdog: 5,
					TimeoutAllocate: 2,
					TimeoutLoop:     1,
					TimeoutMsgwait:  10,
				},
				List: []*cluster.SBDNode{
					{
						ID:     0,
						Name:   "hana01",
						Status: "clear",
					},
					{
						ID:     1,
						Name:   "hana02",
						Status: "clear",
					},
				},
			},
			{
				Device: "/dev/vdb",
				Status: "healthy",
				Dump: cluster.SBDDump{
					Header:          "2.1",
					UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
					Slots:           255,
					SectorSize:      512,
					TimeoutWatchdog: 5,
					TimeoutAllocate: 2,
					TimeoutLoop:     1,
					TimeoutMsgwait:  10,
				},
				List: []*cluster.SBDNode{
					{
						ID:     0,
						Name:   "hana01",
						Status: "clear",
					},
					{
						ID:     1,
						Name:   "hana02",
						Status: "clear",
					},
				},
			},
		},
	}

	suite.EqualExportedValues(expectedSbd, s)
	suite.NoError(err)
}

func (suite *SbdTestSuite) TestNewSBDError() {
	mockCommand := new(mocks.CommandExecutor)
	s, err := cluster.NewSBD(
		mockCommand, "/bin/sbd", helpers.GetFixturePath("discovery/cluster/sbd/sbd_config_no_device"))

	expectedSbd := cluster.SBD{ //nolint
		Config: map[string]string{
			"SBD_OPTS":                "",
			"SBD_PACEMAKER":           "yes",
			"SBD_STARTMODE":           "always",
			"SBD_DELAY_START":         "no",
			"SBD_WATCHDOG_DEV":        "/dev/watchdog",
			"SBD_WATCHDOG_TIMEOUT":    "5",
			"SBD_TIMEOUT_ACTION":      "flush,reboot",
			"SBD_MOVE_TO_ROOT_CGROUP": "auto",
		},
	}

	suite.Equal(expectedSbd, s)
	suite.EqualError(err, "could not find SBD_DEVICE entry in sbd config file")
}

func (suite *SbdTestSuite) TestNewSBDUnhealthyDevices() {
	mockCommand := new(mocks.CommandExecutor)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDumpErr(), errors.New("error"))
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdListErr(), errors.New("error"))
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdb", "dump").Return(mockSbdDumpErr(), errors.New("error"))
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdb", "list").Return(mockSbdListErr(), errors.New("error"))

	s, err := cluster.NewSBD(mockCommand, "/bin/sbd", helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"))

	expectedSbd := cluster.SBD{
		Config: map[string]string{
			"SBD_OPTS":                "",
			"SBD_PACEMAKER":           "yes",
			"SBD_STARTMODE":           "always",
			"SBD_DELAY_START":         "no",
			"SBD_WATCHDOG_DEV":        "/dev/watchdog",
			"SBD_WATCHDOG_TIMEOUT":    "5",
			"SBD_TIMEOUT_ACTION":      "flush,reboot",
			"SBD_MOVE_TO_ROOT_CGROUP": "auto",
			"SBD_DEVICE":              "/dev/vdc;/dev/vdb",
			"AN_INTEGER":              "42",
			"TEST":                    "Value",
			"TEST2":                   "Value2",
		},
		Devices: []*cluster.SBDDevice{
			{
				Device: "/dev/vdc",
				Status: "unhealthy",
				Dump: cluster.SBDDump{
					Header:          "2.1",
					UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
					Slots:           0,
					SectorSize:      0,
					TimeoutWatchdog: 0,
					TimeoutAllocate: 0,
					TimeoutLoop:     0,
					TimeoutMsgwait:  0,
				},
				List: []*cluster.SBDNode{},
			},
			{
				Device: "/dev/vdb",
				Status: "unhealthy",
				Dump: cluster.SBDDump{
					Header:          "2.1",
					UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
					Slots:           0,
					SectorSize:      0,
					TimeoutWatchdog: 0,
					TimeoutAllocate: 0,
					TimeoutLoop:     0,
					TimeoutMsgwait:  0,
				},
				List: []*cluster.SBDNode{},
			},
		},
	}

	suite.EqualExportedValues(expectedSbd, s)
	suite.NoError(err)
}

func (suite *SbdTestSuite) TestNewSBDQuotedDevices() {
	mockCommand := new(mocks.CommandExecutor)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdc", "list").Return(mockSbdList(), nil)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdb", "dump").Return(mockSbdDump(), nil)
	mockCommand.On("Exec", "/bin/sbd", "-d", "/dev/vdb", "list").Return(mockSbdList(), nil)

	s, err := cluster.NewSBD(
		mockCommand, "/bin/sbd", helpers.GetFixturePath("discovery/cluster/sbd/sbd_config_quoted_devices"))

	suite.Equal(len(s.Devices), 2)
	suite.Equal("/dev/vdc", s.Devices[0].Device)
	suite.Equal("/dev/vdb", s.Devices[1].Device)
	suite.NoError(err)
}
