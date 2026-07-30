// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

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
	Totem    map[string]any `json:"totem"`
	Logging  map[string]any `json:"logging"`
	Quorum   map[string]any `json:"quorum"`
	Nodelist map[string]any `json:"nodelist"`
}

func (c *Parser) Parse() (Conf, error) {
	f, err := os.Open(c.configPath)
	if err != nil {
		return Conf{}, fmt.Errorf("error opening corosync.conf: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	config := parseCorosyncConf(scanner)

	err = scanner.Err()
	if err != nil {
		return Conf{}, fmt.Errorf("error reading corosync.conf: %w", err)
	}

	return mapToCorosyncConf(config)
}

func NewCorosyncParser(configPath string) *Parser {
	return &Parser{configPath: configPath}
}

func parseCorosyncConf(scanner *bufio.Scanner) map[string]any {
	root := make(map[string]any)
	stack := []map[string]any{root}

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
			newSection := make(map[string]any)
			current := stack[len(stack)-1]
			current[key] = newSection
			stack = append(stack, newSection)
		} else if before, after, ok := strings.Cut(line, ":"); ok {
			key := strings.TrimSpace(before)
			value := strings.TrimSpace(after)
			current := stack[len(stack)-1]
			current[key] = value
		}
	}

	return root
}

func mapToCorosyncConf(data map[string]any) (Conf, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return Conf{}, fmt.Errorf("error marshalling map to JSON: %w", err)
	}

	var result Conf

	err = json.Unmarshal(b, &result)
	if err != nil {
		return Conf{}, fmt.Errorf("error unmarshalling JSON to struct: %w", err)
	}

	return result, nil
}
