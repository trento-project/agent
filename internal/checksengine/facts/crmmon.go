package facts

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	xmlpath "gopkg.in/xmlpath.v2"
)

const (
	CrmmonFactKey = "crmmon"
)

type crmmonConfigGatherer struct {
}

func NewCrmmonConfigGatherer() *crmmonConfigGatherer {
	return &crmmonConfigGatherer{}
}

func (s *crmmonConfigGatherer) Gather(xmlPaths []string) ([]*Fact, error) {
	var facts []*Fact
	log.Infof("Starting crmmon facts gathering process")

	crmmon, err := exec.Command("crm_mon", "--output-as", "xml").Output()
	if err != nil {
		return facts, err
	}

	crmmonStr := strings.NewReader(string(crmmon))

	root, err := xmlpath.Parse(crmmonStr)
	if err != nil {
		return facts, err
	}

	for _, xPath := range xmlPaths {
		x := xmlpath.MustCompile(xPath)
		fact := &Fact{
			Name:  CrmmonFactKey,
			Key:   xPath,
			Value: fmt.Sprintf("%s not found", xPath),
		}
		if value, ok := x.String(root); ok {
			fact.Value = value
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested crmmon facts gathered")
	return facts, nil
}
