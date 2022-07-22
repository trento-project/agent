package cib

import (
	"encoding/xml"
	"os/exec"

	"github.com/pkg/errors"
)

type Parser struct {
	cibAdminPath string
}

func (p *Parser) Parse() (Root, error) {
	var CIB Root
	cibXML, err := exec.Command(p.cibAdminPath, "--query", "--local").Output() //nolint:gosec
	if err != nil {
		return CIB, errors.Wrap(err, "error while executing cibadmin")
	}

	err = xml.Unmarshal(cibXML, &CIB)
	if err != nil {
		return CIB, errors.Wrap(err, "could not parse cibadmin status from XML")
	}

	return CIB, nil
}

func NewCibAdminParser(cibAdminPath string) *Parser {
	return &Parser{cibAdminPath: cibAdminPath}
}
