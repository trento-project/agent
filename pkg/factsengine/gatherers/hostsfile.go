package gatherers

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	HostsFileGathererName = "hosts"
	HostsFilePath         = "/etc/hosts"
	ipMatchGroup          = "ip"
	hostnamesMatchGroup   = "hostnames"
	hostsParsingRegexp    = `(?m)(?P<` + ipMatchGroup + `>\S+)\s+(?P<` + hostnamesMatchGroup + `>.+)`
)

var (
	hostsEntryCompiled = regexp.MustCompile(hostsParsingRegexp)
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

	HostsFileEntryNotFoundError = entities.FactGatheringError{
		Type:    "hosts-file-value-not-found",
		Message: "requested field value not found in /etc/hosts file",
	}
)

type HostsFileGatherer struct {
	hostsFilePath string
}

func NewDefaultHostsFileGatherer() *HostsFileGatherer {
	return NewHostsFileGatherer(HostsFilePath)
}

func NewHostsFileGatherer(hostsFile string) *HostsFileGatherer {
	return &HostsFileGatherer{hostsFilePath: hostsFile}
}

func (s *HostsFileGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting /etc/hosts file facts gathering process")

	hostsFile, err := readHostsFileByLines(s.hostsFilePath)
	if err != nil {
		return nil, HostsFileError.Wrap(err.Error())
	}

	hostsFileMap, err := hostsFileToMap(hostsFile)
	if err != nil {
		return nil, HostsFileDecodingError.Wrap(err.Error())
	}

	for _, factReq := range factsRequests {
		var fact entities.Fact
		if factReq.Argument == "" {
			fact = entities.NewFactGatheredWithRequest(factReq, hostsFileMap)
			facts = append(facts, fact)
			continue
		}
		if ip, found := hostsFileMap.Value[factReq.Argument]; found {
			fact = entities.NewFactGatheredWithRequest(factReq, ip)
		} else {
			gatheringError := HostsFileEntryNotFoundError.Wrap(factReq.Argument)
			log.Error(gatheringError)
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)
		}
		facts = append(facts, fact)
	}

	log.Infof("Requested /etc/hosts file facts gathered")
	return facts, nil
}

func readHostsFileByLines(filePath string) ([]string, error) {
	hostsFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := hostsFile.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	fileScanner := bufio.NewScanner(hostsFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		scannedLine := fileScanner.Text()
		if strings.HasPrefix(scannedLine, "#") || scannedLine == "" {
			continue
		}
		fileLines = append(fileLines, scannedLine)
	}

	return fileLines, nil
}

func hostsFileToMap(lines []string) (*entities.FactValueMap, error) {
	var hostsFileMap = make(map[string]entities.FactValue)

	var paramsMap = make(map[string]string)

	for _, line := range lines {
		match := hostsEntryCompiled.FindStringSubmatch(line)

		if match == nil {
			return nil, fmt.Errorf("invalid hosts file structure")
		}
		for i, name := range hostsEntryCompiled.SubexpNames() {
			if i > 0 && i <= len(match) {
				paramsMap[name] = match[i]
			}
		}
		hostnames := strings.Fields(paramsMap["hostnames"])

		for _, hostname := range hostnames {
			if ip, found := hostsFileMap[hostname]; found {
				if ipsByHostname, ok := ip.(*entities.FactValueList); ok {
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
