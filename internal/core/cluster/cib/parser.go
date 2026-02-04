package cib

import (
	"encoding/xml"
	"fmt"
	"os/exec"
)

type Parser struct {
	cibAdminPath string
}

func (p *Parser) Parse() (Root, error) {
	var CIB Root
	cibXML, err := exec.Command(p.cibAdminPath, "--query", "--local").Output() //nolint:gosec
	if err != nil {
		return CIB, fmt.Errorf("error while executing cibadmin: %w", err)
	}

	err = xml.Unmarshal(cibXML, &CIB)
	if err != nil {
		return CIB, fmt.Errorf("could not parse cibadmin status from XML: %w", err)
	}

	return CIB, nil
}

func NewCibAdminParser(cibAdminPath string) *Parser {
	return &Parser{cibAdminPath: cibAdminPath}
}
