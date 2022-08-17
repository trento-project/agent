package cluster

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/trento-project/agent/internal/utils"
)

const (
	SBDPath            = "/usr/sbin/sbd"
	SBDConfigPath      = "/etc/sysconfig/sbd"
	SBDStatusUnknown   = "unknown"
	SBDStatusUnhealthy = "unhealthy"
	SBDStatusHealthy   = "healthy"
)

type SBD struct {
	cluster string
	Devices []*SBDDevice           `mapstructure:"devices,omitempty"`
	Config  map[string]interface{} `mapstructure:"config,omitempty"`
}

type SBDDevice struct {
	sbdPath string
	Device  string     `mapstructure:"device,omitempty"`
	Status  string     `mapstructure:"status,omitempty"`
	Dump    SBDDump    `mapstructure:"dump,omitempty"`
	List    []*SBDNode `mapstructure:"list,omitempty"`
}

type SBDDump struct {
	Header          string `mapstructure:"header,omitempty"`
	UUID            string `mapstructure:"uuid,omitempty" json:"Uuid"`
	Slots           int    `mapstructure:"slots,omitempty"`
	SectorSize      int    `mapstructure:"sectorsize,omitempty"`
	TimeoutWatchdog int    `mapstructure:"timeoutwatchdog,omitempty"`
	TimeoutAllocate int    `mapstructure:"timeoutallocate,omitempty"`
	TimeoutLoop     int    `mapstructure:"timeoutloop,omitempty"`
	TimeoutMsgwait  int    `mapstructure:"timeoutmsgwait,omitempty"`
}

type SBDNode struct {
	ID     int    `mapstructure:"id,omitempty" json:"Id"`
	Name   string `mapstructure:"name,omitempty"`
	Status string `mapstructure:"status,omitempty"`
}

var sbdDumpExecCommand = exec.Command //nolint
var sbdListExecCommand = exec.Command //nolint

func NewSBD(cluster, sbdPath, sbdConfigPath string) (SBD, error) {
	var s = SBD{
		cluster: cluster,
		Devices: nil, // TODO check me, no slice of pointers needed
		Config:  map[string]interface{}{},
	}

	c, err := getSBDConfig(sbdConfigPath)
	s.Config = c

	if err != nil {
		return s, err
	} else if _, ok := c["SBD_DEVICE"]; !ok {
		return s, fmt.Errorf("could not find SBD_DEVICE entry in sbd config file")
	}

	sbdDevice, ok := c["SBD_DEVICE"].(string)
	if !ok {
		return s, fmt.Errorf("could not cast sdb device to string, %v", c["SBD_DEVICE"])
	}
	for _, device := range strings.Split(strings.Trim(sbdDevice, "\""), ";") {
		sbdDevice := NewSBDDevice(sbdPath, device)
		err := sbdDevice.LoadDeviceData()
		if err != nil {
			log.Printf("Error getting sbd information: %s", err)
		}
		s.Devices = append(s.Devices, &sbdDevice)
	}

	return s, nil
}

func getSBDConfig(sbdConfigPath string) (map[string]interface{}, error) {
	sbdConfFile, err := os.Open(sbdConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open sbd config file")
	}

	defer sbdConfFile.Close()

	sbdConfigRaw, err := io.ReadAll(sbdConfFile)

	if err != nil {
		return nil, errors.Wrap(err, "could not read sbd config file")
	}

	configMap := utils.FindMatches(`(?m)^(\w+)=(\S[^#\s]*)`, sbdConfigRaw)

	return configMap, nil
}

func NewSBDDevice(sbdPath string, device string) SBDDevice {
	return SBDDevice{ //nolint
		sbdPath: sbdPath,
		Device:  device,
		Status:  SBDStatusUnknown,
	}
}

func (s *SBDDevice) LoadDeviceData() error {
	var sbdErrors []string

	dump, err := sbdDump(s.sbdPath, s.Device)
	s.Dump = dump

	if err != nil {
		s.Status = SBDStatusUnhealthy
		sbdErrors = append(sbdErrors, err.Error())
	} else {
		s.Status = SBDStatusHealthy
	}

	list, err := sbdList(s.sbdPath, s.Device)
	s.List = list

	if err != nil {
		sbdErrors = append(sbdErrors, err.Error())
	}

	if len(sbdErrors) > 0 {
		return fmt.Errorf(strings.Join(sbdErrors, ";"))
	}

	return nil
}

func assignPatternResult(text string, pattern string) string {
	r := regexp.MustCompile(pattern)
	match := r.FindAllStringSubmatch(text, -1)
	if len(match) > 0 {
		return match[0][1]
	}
	// return empty information if pattern is not found
	return ""
}

// Possible output
// ==Dumping header on disk /dev/vdc
// Header version     : 2.1
// UUID               : 541bdcea-16af-44a4-8ab9-6a98602e65ca
// Number of slots    : 255
// Sector size        : 512
// Timeout (watchdog) : 5
// Timeout (allocate) : 2
// Timeout (loop)     : 1
// Timeout (msgwait)  : 10
// ==Header on disk /dev/vdc is dumped
func sbdDump(sbdPath string, device string) (SBDDump, error) {
	var dump = SBDDump{} //nolint

	sbdDump, err := sbdDumpExecCommand(sbdPath, "-d", device, "dump").Output()
	sbdDumpStr := string(sbdDump)

	// FIXME: declarative assignment and error checking on the atoi
	dump.Header = assignPatternResult(sbdDumpStr, `Header version *: (.*)`)
	dump.UUID = assignPatternResult(sbdDumpStr, `UUID *: (.*)`)
	dump.Slots, _ = strconv.Atoi(assignPatternResult(sbdDumpStr, `Number of slots *: (.*)`))
	dump.SectorSize, _ = strconv.Atoi(assignPatternResult(sbdDumpStr, `Sector size *: (.*)`))
	dump.TimeoutWatchdog, _ = strconv.Atoi(assignPatternResult(sbdDumpStr, `Timeout \(watchdog\) *: (.*)`))
	dump.TimeoutAllocate, _ = strconv.Atoi(assignPatternResult(sbdDumpStr, `Timeout \(allocate\) *: (.*)`))
	dump.TimeoutLoop, _ = strconv.Atoi(assignPatternResult(sbdDumpStr, `Timeout \(loop\) *: (.*)`))
	dump.TimeoutMsgwait, _ = strconv.Atoi(assignPatternResult(sbdDumpStr, `Timeout \(msgwait\) *: (.*)`))

	// Sanity check at the end, even in error case the sbd command can output some information
	if err != nil {
		return dump, errors.Wrap(err, "sbd dump command error")
	}

	return dump, nil
}

// Possible output
// 0	hana01	clear
// 1	hana02	clear
func sbdList(sbdPath string, device string) ([]*SBDNode, error) {
	var list = []*SBDNode{}

	output, err := sbdListExecCommand(sbdPath, "-d", device, "list").Output()

	// Loop through sbd list output and find for matches
	r := regexp.MustCompile(`(\d+)\s+(\S+)\s+(\S+)`)
	values := r.FindAllStringSubmatch(string(output), -1)
	for _, match := range values {
		// Continue loop if all the groups are not found
		if len(match) != 4 {
			continue
		}

		id, _ := strconv.Atoi(match[1])
		node := &SBDNode{
			ID:     id,
			Name:   match[2],
			Status: match[3],
		}
		list = append(list, node)
	}

	// Sanity check at the end, even in error case the sbd command can output some information
	if err != nil {
		return list, errors.Wrap(err, "sbd list command error")
	}

	return list, nil
}
