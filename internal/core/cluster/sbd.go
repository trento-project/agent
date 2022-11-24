package cluster

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/go-envparse"
	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/trento-project/agent/pkg/utils"
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
	Devices []*SBDDevice
	Config  map[string]string
}

type SBDDevice struct {
	executor utils.CommandExecutor
	sbdPath  string
	Device   string
	Status   string
	Dump     SBDDump
	List     []*SBDNode
}

type SBDDump struct {
	Header          string
	UUID            string `json:"Uuid"`
	Slots           int
	SectorSize      int
	TimeoutWatchdog int
	TimeoutAllocate int
	TimeoutLoop     int
	TimeoutMsgwait  int
}

type SBDNode struct {
	ID     int `json:"Id"`
	Name   string
	Status string
}

func NewSBD(executor utils.CommandExecutor, cluster, sbdPath, sbdConfigPath string) (SBD, error) {
	var s = SBD{
		cluster: cluster,
		Devices: nil, // TODO check me, no slice of pointers needed
		Config:  map[string]string{},
	}

	c, err := LoadSbdConfig(sbdConfigPath)
	s.Config = c

	if err != nil {
		return s, err
	} else if _, ok := c["SBD_DEVICE"]; !ok {
		return s, fmt.Errorf("could not find SBD_DEVICE entry in sbd config file")
	}

	sbdDevice, ok := c["SBD_DEVICE"]
	if !ok {
		return s, fmt.Errorf("could not cast sbd device to string, %v", c["SBD_DEVICE"])
	}
	for _, device := range strings.Split(strings.Trim(sbdDevice, "\""), ";") {
		sbdDevice := NewSBDDevice(executor, sbdPath, device)
		err := sbdDevice.LoadDeviceData()
		if err != nil {
			log.Printf("Error getting sbd information: %s", err)
		}
		s.Devices = append(s.Devices, &sbdDevice)
	}

	return s, nil
}

func LoadSbdConfig(sbdConfigPath string) (map[string]string, error) {
	sbdConfigFile, err := os.Open(sbdConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open sbd config file")
	}

	defer func() {
		err := sbdConfigFile.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	conf, err := envparse.Parse(sbdConfigFile)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse sbd config file")
	}

	return conf, nil
}

func NewSBDDevice(executor utils.CommandExecutor, sbdPath, device string) SBDDevice {
	return SBDDevice{ //nolint
		executor: executor,
		sbdPath:  sbdPath,
		Device:   device,
		Status:   SBDStatusUnknown,
	}
}

func (s *SBDDevice) LoadDeviceData() error {
	var sbdErrors []string

	dump, err := sbdDump(s.executor, s.sbdPath, s.Device)
	s.Dump = dump

	if err != nil {
		s.Status = SBDStatusUnhealthy
		sbdErrors = append(sbdErrors, err.Error())
	} else {
		s.Status = SBDStatusHealthy
	}

	list, err := sbdList(s.executor, s.sbdPath, s.Device)
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
func sbdDump(executor utils.CommandExecutor, sbdPath string, device string) (SBDDump, error) {
	sbdDumpOutput, dumpErr := executor.Exec(sbdPath, "-d", device, "dump")
	sbdDumpStr := string(sbdDumpOutput)

	header := assignPatternResult(sbdDumpStr, `Header version *: (.*)`)
	uuid := assignPatternResult(sbdDumpStr, `UUID *: (.*)`)
	slots, err := strconv.Atoi(assignPatternResult(sbdDumpStr, `Number of slots *: (.*)`))
	if err != nil {
		log.Error("Error parsing Number of slots value as integer")
	}
	sectorSize, err := strconv.Atoi(assignPatternResult(sbdDumpStr, `Sector size *: (.*)`))
	if err != nil {
		log.Error("Error parsing Sector size value as integer")
	}
	timeoutWatchdog, err := strconv.Atoi(assignPatternResult(sbdDumpStr, `Timeout \(watchdog\) *: (.*)`))
	if err != nil {
		log.Error("Error parsing Timeout watchdog value as integer")
	}
	timeoutAllocate, err := strconv.Atoi(assignPatternResult(sbdDumpStr, `Timeout \(allocate\) *: (.*)`))
	if err != nil {
		log.Error("Error parsing Tiemout allocate value as integer")
	}
	timeoutLoop, err := strconv.Atoi(assignPatternResult(sbdDumpStr, `Timeout \(loop\) *: (.*)`))
	if err != nil {
		log.Error("Error parsing Timeout loop value as integer")
	}
	timeoutMsgwait, err := strconv.Atoi(assignPatternResult(sbdDumpStr, `Timeout \(msgwait\) *: (.*)`))
	if err != nil {
		log.Error("Error parsing Timeout msgwait value as integer")
	}

	sbdDump := SBDDump{
		Header:          header,
		UUID:            uuid,
		Slots:           slots,
		SectorSize:      sectorSize,
		TimeoutWatchdog: timeoutWatchdog,
		TimeoutAllocate: timeoutAllocate,
		TimeoutLoop:     timeoutLoop,
		TimeoutMsgwait:  timeoutMsgwait,
	}

	if dumpErr != nil {
		return sbdDump, errors.Wrap(dumpErr, "sbd dump command error")
	}

	return sbdDump, nil
}

// Possible output
// 0	hana01	clear
// 1	hana02	clear
func sbdList(executor utils.CommandExecutor, sbdPath, device string) ([]*SBDNode, error) {
	var list = []*SBDNode{}

	output, err := executor.Exec(sbdPath, "-d", device, "list")

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
