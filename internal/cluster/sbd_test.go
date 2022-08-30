package cluster

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SbdTestSuite struct {
	suite.Suite
}

func TestSbdTestSuite(t *testing.T) {
	suite.Run(t, new(SbdTestSuite))
}

func mockSbdDump(command string, args ...string) *exec.Cmd {
	cmd := `==Dumping header on disk /dev/vdc
Header version     : 2.1
UUID               : 541bdcea-16af-44a4-8ab9-6a98602e65ca
Number of slots    : 255
Sector size        : 512
Timeout (watchdog) : 5
Timeout (allocate) : 2
Timeout (loop)     : 1
Timeout (msgwait)  : 10
==Header on disk /dev/vdc is dumped`
	return exec.Command("echo", cmd)
}

func mockSbdDumpErr(command string, args ...string) *exec.Cmd {
	cmd := `==Dumping header on disk /dev/vdc
Header version     : 2.1
UUID               : 541bdcea-16af-44a4-8ab9-6a98602e65ca
==Number of slots on disk /dev/vdb NOT dumped
sbd failed; please check the logs.`

	script := fmt.Sprintf("echo \"%s\" && exit 1", cmd)

	return exec.Command("bash", "-c", script)
}

func mockSbdList(command string, args ...string) *exec.Cmd {
	cmd := `0	hana01	clear
1	hana02	clear`
	return exec.Command("echo", cmd)
}

func mockSbdListErr(command string, args ...string) *exec.Cmd {
	cmd := `== disk /dev/vdxx unreadable!
sbd failed; please check the logs.`

	script := fmt.Sprintf("echo \"%s\" && exit 1", cmd)

	return exec.Command("bash", "-c", script)
}

func (suite *SbdTestSuite) TestSbdDump() {
	sbdDumpExecCommand = mockSbdDump

	dump, err := sbdDump("/bin/sbd", "/dev/vdc")

	expectedDump := SBDDump{
		Header:          "2.1",
		UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
		Slots:           255,
		SectorSize:      512,
		TimeoutWatchdog: 5,
		TimeoutAllocate: 2,
		TimeoutLoop:     1,
		TimeoutMsgwait:  10,
	}

	suite.Equal(expectedDump, dump)
	suite.NoError(err)
}

func (suite *SbdTestSuite) TestSbdDumpError() {
	sbdDumpExecCommand = mockSbdDumpErr

	dump, err := sbdDump("/bin/sbd", "/dev/vdc")

	expectedDump := SBDDump{
		Header:          "2.1",
		UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
		Slots:           0,
		SectorSize:      0,
		TimeoutWatchdog: 0,
		TimeoutAllocate: 0,
		TimeoutLoop:     0,
		TimeoutMsgwait:  0,
	}

	suite.Equal(expectedDump, dump)
	suite.EqualError(err, "sbd dump command error: exit status 1")
}

func (suite *SbdTestSuite) TestSbdList() {
	sbdListExecCommand = mockSbdList

	list, err := sbdList("/bin/sbd", "/dev/vdc")

	expectedList := []*SBDNode{
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
	}

	suite.Equal(expectedList, list)
	suite.NoError(err)
}

func (suite *SbdTestSuite) TestSbdListError() {
	sbdListExecCommand = mockSbdListErr

	list, err := sbdList("/bin/sbd", "/dev/vdc")

	expectedList := []*SBDNode{}

	suite.Equal(expectedList, list)
	suite.EqualError(err, "sbd list command error: exit status 1")
}

func (suite *SbdTestSuite) TestLoadDeviceData() {
	s := NewSBDDevice("/bin/sbd", "/dev/vdc")

	sbdDumpExecCommand = mockSbdDump
	sbdListExecCommand = mockSbdList

	err := s.LoadDeviceData()

	expectedDevice := NewSBDDevice("/bin/sbd", "/dev/vdc")
	expectedDevice.Status = "healthy"
	expectedDevice.Dump = SBDDump{
		Header:          "2.1",
		UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
		Slots:           255,
		SectorSize:      512,
		TimeoutWatchdog: 5,
		TimeoutAllocate: 2,
		TimeoutLoop:     1,
		TimeoutMsgwait:  10,
	}
	expectedDevice.List = []*SBDNode{
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
	}

	suite.Equal(expectedDevice, s)
	suite.NoError(err)
}

func (suite *SbdTestSuite) TestLoadDeviceDataDumpError() {
	s := NewSBDDevice("/bin/sbdErr", "/dev/vdc")

	sbdDumpExecCommand = mockSbdDumpErr

	err := s.LoadDeviceData()

	expectedDevice := NewSBDDevice("/bin/sbdErr", "/dev/vdc")
	expectedDevice.Status = "unhealthy"

	expectedDevice.Dump = SBDDump{
		Header:          "2.1",
		UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
		Slots:           0,
		SectorSize:      0,
		TimeoutWatchdog: 0,
		TimeoutAllocate: 0,
		TimeoutLoop:     0,
		TimeoutMsgwait:  0,
	}

	expectedDevice.List = []*SBDNode{
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
	}

	suite.Equal(expectedDevice, s)
	suite.EqualError(err, "sbd dump command error: exit status 1")
}

func (suite *SbdTestSuite) TestLoadDeviceDataListError() {
	s := NewSBDDevice("/bin/sbdErr", "/dev/vdc")

	sbdDumpExecCommand = mockSbdDump
	sbdListExecCommand = mockSbdListErr

	err := s.LoadDeviceData()

	expectedDevice := NewSBDDevice("/bin/sbdErr", "/dev/vdc")
	expectedDevice.Status = "healthy"
	expectedDevice.Dump = SBDDump{
		Header:          "2.1",
		UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
		Slots:           255,
		SectorSize:      512,
		TimeoutWatchdog: 5,
		TimeoutAllocate: 2,
		TimeoutLoop:     1,
		TimeoutMsgwait:  10,
	}

	expectedDevice.List = []*SBDNode{}

	suite.Equal(expectedDevice, s)
	suite.EqualError(err, "sbd list command error: exit status 1")
}

func (suite *SbdTestSuite) TestLoadDeviceDataError() {
	s := NewSBDDevice("/bin/sbdErr", "/dev/vdc")

	sbdDumpExecCommand = mockSbdDumpErr
	sbdListExecCommand = mockSbdListErr

	err := s.LoadDeviceData()

	expectedDevice := NewSBDDevice("/bin/sbdErr", "/dev/vdc")
	expectedDevice.Status = "unhealthy"

	expectedDevice.Dump = SBDDump{
		Header:          "2.1",
		UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
		Slots:           0,
		SectorSize:      0,
		TimeoutWatchdog: 0,
		TimeoutAllocate: 0,
		TimeoutLoop:     0,
		TimeoutMsgwait:  0,
	}

	expectedDevice.List = []*SBDNode{}

	suite.Equal(expectedDevice, s)
	suite.EqualError(err, "sbd dump command error: exit status 1;sbd list command error: exit status 1")
}

func (suite *SbdTestSuite) TestGetSBDConfig() {
	sbdConfig, err := getSBDConfig("../../test/sbd_config")

	expectedConfig := map[string]interface{}{
		"SBD_PACEMAKER":           "yes",
		"SBD_STARTMODE":           "always",
		"SBD_DELAY_START":         "no",
		"SBD_WATCHDOG_DEV":        "/dev/watchdog",
		"SBD_WATCHDOG_TIMEOUT":    "5",
		"SBD_TIMEOUT_ACTION":      "flush,reboot",
		"SBD_MOVE_TO_ROOT_CGROUP": "auto",
		"SBD_DEVICE":              "/dev/vdc;/dev/vdb",
		"TEST":                    "Value",
		"TEST2":                   "Value2",
	}

	suite.Equal(expectedConfig, sbdConfig)
	suite.NoError(err)
}

func (suite *SbdTestSuite) TestGetSBDConfigError() {
	sbdConfig, err := getSBDConfig("notexist")

	expectedConfig := map[string]interface{}(nil)

	suite.Equal(expectedConfig, sbdConfig)
	suite.EqualError(err, "could not open sbd config file: open notexist: no such file or directory")
}

func (suite *SbdTestSuite) TestNewSBD() {
	sbdDumpExecCommand = mockSbdDump
	sbdListExecCommand = mockSbdList

	s, err := NewSBD("mycluster", "/bin/sbd", "../../test/sbd_config")

	expectedSbd := SBD{
		cluster: "mycluster",
		Config: map[string]interface{}{
			"SBD_PACEMAKER":           "yes",
			"SBD_STARTMODE":           "always",
			"SBD_DELAY_START":         "no",
			"SBD_WATCHDOG_DEV":        "/dev/watchdog",
			"SBD_WATCHDOG_TIMEOUT":    "5",
			"SBD_TIMEOUT_ACTION":      "flush,reboot",
			"SBD_MOVE_TO_ROOT_CGROUP": "auto",
			"SBD_DEVICE":              "/dev/vdc;/dev/vdb",
			"TEST":                    "Value",
			"TEST2":                   "Value2",
		},
		Devices: []*SBDDevice{
			{
				sbdPath: "/bin/sbd",
				Device:  "/dev/vdc",
				Status:  "healthy",
				Dump: SBDDump{
					Header:          "2.1",
					UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
					Slots:           255,
					SectorSize:      512,
					TimeoutWatchdog: 5,
					TimeoutAllocate: 2,
					TimeoutLoop:     1,
					TimeoutMsgwait:  10,
				},
				List: []*SBDNode{
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
				sbdPath: "/bin/sbd",
				Device:  "/dev/vdb",
				Status:  "healthy",
				Dump: SBDDump{
					Header:          "2.1",
					UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
					Slots:           255,
					SectorSize:      512,
					TimeoutWatchdog: 5,
					TimeoutAllocate: 2,
					TimeoutLoop:     1,
					TimeoutMsgwait:  10,
				},
				List: []*SBDNode{
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

	suite.Equal(expectedSbd, s)
	suite.NoError(err)
}

func (suite *SbdTestSuite) TestNewSBDError() {
	s, err := NewSBD("mycluster", "/bin/sbd", "../../test/sbd_config_no_device")

	expectedSbd := SBD{ //nolint
		cluster: "mycluster",
		Config: map[string]interface{}{
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
	sbdDumpExecCommand = mockSbdDumpErr
	sbdListExecCommand = mockSbdListErr

	s, err := NewSBD("mycluster", "/bin/sbd", "../../test/sbd_config")

	expectedSbd := SBD{
		cluster: "mycluster",
		Config: map[string]interface{}{
			"SBD_PACEMAKER":           "yes",
			"SBD_STARTMODE":           "always",
			"SBD_DELAY_START":         "no",
			"SBD_WATCHDOG_DEV":        "/dev/watchdog",
			"SBD_WATCHDOG_TIMEOUT":    "5",
			"SBD_TIMEOUT_ACTION":      "flush,reboot",
			"SBD_MOVE_TO_ROOT_CGROUP": "auto",
			"SBD_DEVICE":              "/dev/vdc;/dev/vdb",
			"TEST":                    "Value",
			"TEST2":                   "Value2",
		},
		Devices: []*SBDDevice{
			{
				sbdPath: "/bin/sbd",
				Device:  "/dev/vdc",
				Status:  "unhealthy",
				Dump: SBDDump{
					Header:          "2.1",
					UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
					Slots:           0,
					SectorSize:      0,
					TimeoutWatchdog: 0,
					TimeoutAllocate: 0,
					TimeoutLoop:     0,
					TimeoutMsgwait:  0,
				},
				List: []*SBDNode{},
			},
			{
				sbdPath: "/bin/sbd",
				Device:  "/dev/vdb",
				Status:  "unhealthy",
				Dump: SBDDump{
					Header:          "2.1",
					UUID:            "541bdcea-16af-44a4-8ab9-6a98602e65ca",
					Slots:           0,
					SectorSize:      0,
					TimeoutWatchdog: 0,
					TimeoutAllocate: 0,
					TimeoutLoop:     0,
					TimeoutMsgwait:  0,
				},
				List: []*SBDNode{},
			},
		},
	}

	suite.Equal(expectedSbd, s)
	suite.NoError(err)
}

func (suite *SbdTestSuite) TestNewSBDQuotedDevices() {
	sbdDumpExecCommand = mockSbdDump
	sbdListExecCommand = mockSbdList

	s, err := NewSBD("mycluster", "/bin/sbd", "../../test/sbd_config_quoted_devices")

	suite.Equal(len(s.Devices), 2)
	suite.Equal("/dev/vdc", s.Devices[0].Device)
	suite.Equal("/dev/vdb", s.Devices[1].Device)
	suite.NoError(err)
}
