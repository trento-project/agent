package gatherers

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	CorosyncFactKey  = "corosync.conf"
	CorosyncConfPath = "/etc/corosync/corosync.conf"
)

var fileSystem = afero.NewOsFs()

type corosyncConfGatherer struct {
}

func NewCorosyncConfGatherer() *corosyncConfGatherer {
	return &corosyncConfGatherer{}
}

func (s *corosyncConfGatherer) Gather(factsRequests []*FactRequest) ([]*Fact, error) {
	var facts []*Fact
	log.Infof("Starting corosync.conf file facts gathering process")

	corosyncConfile, err := readCorosyncConfFileByLines(CorosyncConfPath)
	if err != nil {
		return facts, err
	}

	corosycnMap, err := corosyncConfToMap(corosyncConfile)
	if err != nil {
		return facts, err
	}

	for _, factReq := range factsRequests {
		fact := &Fact{
			Name:  factReq.Name,
			Value: getValue(corosycnMap, strings.Split(factReq.Argument, ".")),
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested corosync.conf file facts gathered")
	return facts, nil
}

func readCorosyncConfFileByLines(filePath string) ([]string, error) {
	corosyncConfFile, err := fileSystem.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open corosync.conf file %s", err)
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

func corosyncConfToMap(lines []string) (map[string]interface{}, error) {
	var corosyncMap = make(map[string]interface{})
	var sectionStartPattern = `^\s*(\w+)\s*{.*`
	var sectionEndPattern = `^\s*}.*`
	var valuePattern = `^\s*(\w+)\s*:\s*(\S+).*`
	var sections int = 0

	sectionStartPatternCompiled := regexp.MustCompile(sectionStartPattern)
	sectionEndPatternCompiled := regexp.MustCompile(sectionEndPattern)
	valuePatternCompiled := regexp.MustCompile(valuePattern)

	for index, line := range lines {
		if start := sectionStartPatternCompiled.FindStringSubmatch(line); start != nil {
			if sections == 0 {
				children, _ := corosyncConfToMap(lines[index+1:])
				if value, found := corosyncMap[start[1]]; found {
					newList := []interface{}{value}
					corosyncMap[start[1]] = append(newList, children)
				} else {
					corosyncMap[start[1]] = children
				}
			}
			sections++
			continue
		}

		if end := sectionEndPatternCompiled.FindStringSubmatch(line); end != nil {
			if sections == 0 {
				return corosyncMap, nil
			}
			sections--
			continue
		}

		if value := valuePatternCompiled.FindStringSubmatch(line); value != nil && sections == 0 {
			corosyncMap[value[1]] = value[2]
			continue
		}
	}

	if sections != 0 {
		return corosyncMap, fmt.Errorf("invalid corosync file structure. some section is not closed properly")
	}

	return corosyncMap, nil
}

func getValue(corosyncMap map[string]interface{}, values []string) interface{} {
	if len(values) == 0 {
		return corosyncMap
	}

	if value, found := corosyncMap[values[0]]; found {
		switch value.(type) {
		case map[string]interface{}:
			return getValue(value.(map[string]interface{}), values[1:])
		case []interface{}:
			// Requested value is the whole list of elements
			if len(values) < 2 {
				return value.([]interface{})
			}
			listIndex, err := strconv.Atoi(values[1])
			if err != nil {
				return fmt.Sprintf("%s value is a list. Must be followed by an integer value", values[0])
			}
			return getValue(value.([]interface{})[listIndex].(map[string]interface{}), values[2:])
		default:
			return value
		}
	} else {
		return nil
	}

}
