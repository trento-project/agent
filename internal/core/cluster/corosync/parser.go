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

type CorosyncConf struct {
	Totem    map[string]interface{} `json:"totem"`
	Logging  map[string]interface{} `json:"logging"`
	Quorum   map[string]interface{} `json:"quorum"`
	Nodelist map[string]interface{} `json:"nodelist"`
}

func (c *Parser) Parse() (CorosyncConf, error) {
	f, err := os.Open(c.configPath)
	if err != nil {
		return CorosyncConf{}, fmt.Errorf("error opening corosync.conf: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	config := parseSection(scanner)

	if err := scanner.Err(); err != nil {
		return CorosyncConf{}, fmt.Errorf("error reading corosync.conf: %w", err)
	}

	return mapToCorosyncConf(config)

}

func NewCorosyncParser(configPath string) *Parser {
	return &Parser{configPath: configPath}
}

// Recursively parses a section.
func parseSection(scanner *bufio.Scanner) map[string]interface{} {
	config := make(map[string]interface{})
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// end of section
		if line == "}" {
			break
		}

		// if line ends with '{', it's a nested section
		// otherwise, it's a key-value pair
		if strings.HasSuffix(line, "{") {
			key := strings.TrimSpace(line[:len(line)-1])
			config[key] = parseSection(scanner)
		} else if sep := strings.Index(line, ":"); sep > -1 {
			key := strings.TrimSpace(line[:sep])
			value := strings.TrimSpace(line[sep+1:])
			config[key] = value
		}
	}

	return config
}

func mapToCorosyncConf(data map[string]interface{}) (CorosyncConf, error) {

	b, err := json.Marshal(data)
	if err != nil {
		return CorosyncConf{}, fmt.Errorf("error marshalling map to JSON: %w", err)
	}

	var result CorosyncConf
	if err := json.Unmarshal(b, &result); err != nil {
		return CorosyncConf{}, fmt.Errorf("error unmarshalling JSON to struct: %w", err)
	}

	return result, nil

}
