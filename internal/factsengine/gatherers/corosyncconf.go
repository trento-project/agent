package gatherers

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	CorosyncConfGathererName = "corosync.conf"
	CorosyncConfPath         = "/etc/corosync/corosync.conf"
)

var (
	sectionStartPatternCompiled = regexp.MustCompile(`^\s*(\w+)\s*{.*`)
	sectionEndPatternCompiled   = regexp.MustCompile(`^\s*}.*`)
	valuePatternCompiled        = regexp.MustCompile(`^\s*(\w+)\s*:\s*(\S+).*`)
)

// nolint:gochecknoglobals
var (
	CorosyncConfFileError = entities.FactGatheringError{
		Type:    "corosync-conf-file-error",
		Message: "error reading corosync.conf file",
	}

	CorosyncConfDecodingError = entities.FactGatheringError{
		Type:    "corosync-conf-decoding-error",
		Message: "error decoding corosync.conf file",
	}
)

type CorosyncConfGatherer struct {
	configFile string
}

func NewDefaultCorosyncConfGatherer() *CorosyncConfGatherer {
	return NewCorosyncConfGatherer(CorosyncConfPath)
}

func NewCorosyncConfGatherer(configFile string) *CorosyncConfGatherer {
	return &CorosyncConfGatherer{
		configFile,
	}
}

func (s *CorosyncConfGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting corosync.conf file facts gathering process")

	corosyncConfile, err := readCorosyncConfFileByLines(s.configFile)
	if err != nil {
		return nil, CorosyncConfFileError.Wrap(err.Error())
	}

	elementsToList := map[string]bool{"interface": true, "node": true}

	corosyncMap, err := corosyncConfToMap(corosyncConfile, elementsToList)
	if err != nil {
		return nil, CorosyncConfDecodingError.Wrap(err.Error())
	}

	for _, factReq := range factsRequests {
		var fact entities.Fact

		if value, err := corosyncMap.GetValue(factReq.Argument); err == nil {
			fact = entities.NewFactGatheredWithRequest(factReq, value)

		} else {
			log.Error(err)
			fact = entities.NewFactGatheredWithError(factReq, err)
		}
		facts = append(facts, fact)
	}

	log.Infof("Requested corosync.conf file facts gathered")
	return facts, nil
}

func readCorosyncConfFileByLines(filePath string) ([]string, error) {
	corosyncConfFile, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open corosync.conf file")
	}

	defer corosyncConfFile.Close()

	fileScanner := bufio.NewScanner(corosyncConfFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}

	return fileLines, nil
}

func corosyncConfToMap(lines []string, elementsToList map[string]bool) (*entities.FactValueMap, error) {
	var cm = make(map[string]entities.FactValue)
	var sections int

	for index, line := range lines {
		if start := sectionStartPatternCompiled.FindStringSubmatch(line); start != nil {
			if sections == 0 {
				sectionKey := start[1]
				_, found := cm[sectionKey]
				if !found && elementsToList[sectionKey] {
					cm[sectionKey] = &entities.FactValueList{Value: []entities.FactValue{}}
				}

				children, _ := corosyncConfToMap(lines[index+1:], elementsToList)

				if elementsToList[sectionKey] {
					factList, ok := cm[sectionKey].(*entities.FactValueList)
					if !ok {
						return nil, fmt.Errorf("error asserting to list type for key: %s", sectionKey)
					}
					factList.AppendValue(children)
				} else {
					cm[sectionKey] = children
				}
			}
			sections++
			continue
		}

		if end := sectionEndPatternCompiled.FindStringSubmatch(line); end != nil {
			if sections == 0 {
				return &entities.FactValueMap{
					Value: cm,
				}, nil
			}
			sections--
			continue
		}

		if value := valuePatternCompiled.FindStringSubmatch(line); value != nil && sections == 0 {
			cm[value[1]] = entities.ParseStringToFactValue(value[2])
			continue
		}
	}

	corosyncMap := &entities.FactValueMap{
		Value: cm,
	}

	if sections != 0 {
		return corosyncMap, fmt.Errorf("invalid corosync file structure. some section is not closed properly")
	}

	return corosyncMap, nil
}
