package crmmon

import (
	"encoding/xml"
	"fmt"
	"os/exec"
)

type Parser struct {
	crmMonPath string
}

func (c *Parser) Parse() (Root, error) {
	var crmMon Root

	crmMonXML, err := exec.Command(c.crmMonPath, "-X", "--inactive").Output() //nolint:gosec
	if err != nil {
		return crmMon, fmt.Errorf("error while executing crm_mon: %w", err)
	}

	err = xml.Unmarshal(crmMonXML, &crmMon)
	if err != nil {
		return crmMon, fmt.Errorf("error while parsing crm_mon XML output: %w", err)
	}

	return crmMon, nil
}

func NewCrmMonParser(crmMonPath string) *Parser {
	return &Parser{crmMonPath: crmMonPath}
}
