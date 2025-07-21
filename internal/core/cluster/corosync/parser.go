package corosync

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Parser struct {
	configPath string
}

type Conf struct {
	Totem    map[string]interface{} `json:"totem"`
	Logging  map[string]interface{} `json:"logging"`
	Quorum   map[string]interface{} `json:"quorum"`
	Nodelist map[string]interface{} `json:"nodelist"`
}

func (c *Parser) Parse() (Conf, error) {
	f, err := os.Open(c.configPath)
	if err != nil {
		return Conf{}, fmt.Errorf("error opening corosync.conf: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	config := parseCorisyncConf(scanner)

	if err := scanner.Err(); err != nil {
		return Conf{}, fmt.Errorf("error reading corosync.conf: %w", err)
	}

	return mapToCorosyncConf(config)

}

func NewCorosyncParser(configPath string) *Parser {
	return &Parser{configPath: configPath}
}

func parseCorisyncConf(scanner *bufio.Scanner) map[string]interface{} {
	root := make(map[string]interface{})
	stack := []map[string]interface{}{root}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// end of section
		if line == "}" {
			// pop the current section from the stack
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
			continue
		}
		// if line ends with '{', it's a nested section
		// otherwise, it's a key-value pair
		if strings.HasSuffix(line, "{") {
			key := strings.TrimSpace(line[:len(line)-1])
			newSection := make(map[string]interface{})
			current := stack[len(stack)-1]
			current[key] = newSection
			stack = append(stack, newSection)
		} else if sep := strings.Index(line, ":"); sep > -1 {
			key := strings.TrimSpace(line[:sep])
			value := strings.TrimSpace(line[sep+1:])
			current := stack[len(stack)-1]
			current[key] = value
		}
	}
	return root
}

func mapToCorosyncConf(data map[string]interface{}) (Conf, error) {

	b, err := json.Marshal(data)
	if err != nil {
		return Conf{}, fmt.Errorf("error marshalling map to JSON: %w", err)
	}

	var result Conf
	if err := json.Unmarshal(b, &result); err != nil {
		return Conf{}, fmt.Errorf("error unmarshalling JSON to struct: %w", err)
	}

	return result, nil

}
