package crmmon

import (
	"encoding/xml"
	"os/exec"

	"github.com/pkg/errors"
)

type Parser struct {
	crmMonPath string
}

func (c *Parser) Parse() (Root, error) {
	var crmMon Root

	crmMonXML, err := exec.Command(c.crmMonPath, "-X", "--inactive").Output() //nolint:gosec
	if err != nil {
		return crmMon, errors.Wrap(err, "error while executing crm_mon")
	}

	err = xml.Unmarshal(crmMonXML, &crmMon)
	if err != nil {
		return crmMon, errors.Wrap(err, "error while parsing crm_mon XML output")
	}

	return crmMon, nil
}

func NewCrmMonParser(crmMonPath string) *Parser {
	return &Parser{crmMonPath: crmMonPath}
}
