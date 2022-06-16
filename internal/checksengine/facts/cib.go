package facts

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	xmlpath "gopkg.in/xmlpath.v2"
)

const (
	CibFactKey = "cib"
)

type cibConfigGatherer struct {
}

func NewcibConfigGatherer() *cibConfigGatherer {
	return &cibConfigGatherer{}
}

func (s *cibConfigGatherer) Gather(xmlPaths []string) ([]*Fact, error) {
	var facts []*Fact
	log.Infof("Starting CIB facts gathering process")

	cib, err := exec.Command("cibadmin", "--query", "--local").Output()
	if err != nil {
		return facts, err
	}

	cibStr := strings.NewReader(string(cib))

	root, err := xmlpath.Parse(cibStr)
	if err != nil {
		return facts, err
	}

	for _, xPath := range xmlPaths {
		x := xmlpath.MustCompile(xPath)
		fact := &Fact{
			Name:  CibFactKey,
			Key:   xPath,
			Value: fmt.Sprintf("%s not found", xPath),
		}
		if value, ok := x.String(root); ok {
			fact.Value = value
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested CIB facts gathered")
	return facts, nil
}
