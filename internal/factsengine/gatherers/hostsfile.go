package gatherers

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

const (
	HostsFileFactKey    = "hosts"
	HostsFilePath       = "/etc/hosts"
	ipMatchGroup        = "ip"
	hostnamesMatchGroup = "hostnames"
	parsingRegexp       = `(?m)(?P<` + ipMatchGroup + `>\S+)\s+(?P<` + hostnamesMatchGroup + `>.+)`
)

var (
	HostsEntryCompiled = regexp.MustCompile(parsingRegexp)
)

// nolint:gochecknoglobals
var (
	HostsFileError = entities.FactGatheringError{
		Type:    "hosts-file-error",
		Message: "error reading /etc/hosts file",
	}

	HostsFileDecodingError = entities.FactGatheringError{
		Type:    "hosts-file-decoding-error",
		Message: "error decoding /etc/hosts file",
	}

	HostsEntryNotFoundError = entities.FactGatheringError{
		Type:    "hosts-file-value-not-found",
		Message: "requested field value not found in /etc/hosts file",
	}
)

type HostsFileGatherer struct {
	hostsFile string
}

func NewDefaultHostsFileGatherer() *HostsFileGatherer {
	return NewHostsFileGatherer(HostsFilePath)
}

func NewHostsFileGatherer(hostsFile string) *HostsFileGatherer {
	return &HostsFileGatherer{hostsFile: hostsFile}
}

func (s *HostsFileGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting /etc/hosts file facts gathering process")

	hostsFile, err := readHostsFileByLines(s.hostsFile)
	if err != nil {
		return nil, HostsFileError.Wrap(err.Error())
	}

	hostsFileMap, err := hostsFileToMap(hostsFile)
	if err != nil {
		return nil, HostsFileDecodingError.Wrap(err.Error())
	}

	for _, factReq := range factsRequests {
		var fact entities.Fact
		var found bool

		for hostname, ip := range hostsFileMap.Value {
			if hostname == factReq.Argument {
				fact = entities.NewFactGatheredWithRequest(factReq, ip)
				facts = append(facts, fact)
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New(HostsEntryNotFoundError.Error())
		}
	}

	log.Infof("Requested /etc/hosts file facts gathered")
	return facts, nil
}

func readHostsFileByLines(filePath string) ([]string, error) {
	hostsFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer hostsFile.Close()

	fileScanner := bufio.NewScanner(hostsFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		if strings.HasPrefix(fileScanner.Text(), "#") || fileScanner.Text() == "" {
			continue
		}
		fileLines = append(fileLines, fileScanner.Text())
	}

	return fileLines, nil
}

func hostsFileToMap(lines []string) (*entities.FactValueMap, error) {
	var hostsFileMap = make(map[string]entities.FactValue)

	var paramsMap = make(map[string]string)

	for _, line := range lines {
		match := HostsEntryCompiled.FindStringSubmatch(line)

		if match == nil {
			return nil, fmt.Errorf("invalid hosts file structure")
		}
		for i, name := range HostsEntryCompiled.SubexpNames() {
			if i > 0 && i <= len(match) {
				paramsMap[name] = match[i]
			}
		}
		hostnames := strings.Fields(paramsMap["hostnames"])

		for _, hostname := range hostnames {
			if i, found := hostsFileMap[hostname]; found {
				if ipsByHostname, ok := i.(*entities.FactValueList); ok {
					ipsByHostname.Value = append(ipsByHostname.Value, &entities.FactValueString{Value: paramsMap["ip"]})
				} else {
					return nil, fmt.Errorf("casting error while mapping ips to hosts")
				}

			} else {
				hostsFileMap[hostname] = &entities.FactValueList{Value: []entities.FactValue{
					&entities.FactValueString{Value: paramsMap["ip"]},
				}}
			}
		}
	}

	return &entities.FactValueMap{Value: hostsFileMap}, nil
}
