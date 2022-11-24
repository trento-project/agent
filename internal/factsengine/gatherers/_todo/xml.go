package gatherers

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	xmlpath "gopkg.in/xmlpath.v2"
)

func GatherFromXML(xmlContent string, factsRequests []entities.FactRequest) ([]entities.FactsGatheredItem, error) {
	facts := []entities.FactsGatheredItem{}
	r := strings.NewReader(xmlContent)

	root, err := xmlpath.Parse(r)
	if err != nil {
		return facts, err
	}

	for _, factReq := range factsRequests {
		x, err := xmlpath.Compile(factReq.Argument)
		if err != nil {
			log.Errorf("Error compiling xpath: %s", factReq.Argument)
			return nil, err
		}

		value, ok := x.String(root)
		if !ok {
			// TODO: Decide together with Wanda how to deal with errors. `err` field in the fact result?
			log.Errorf("Value with provided xpath not found: %s", factReq.Argument)
		}

		fact := entities.NewFactGatheredWithRequest(factReq, value)
		facts = append(facts, fact)
	}

	return facts, nil
}
