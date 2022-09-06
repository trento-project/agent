package gatherers

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
	xmlpath "gopkg.in/xmlpath.v2"
)

var (
	XmlCompileError = entities.FactGatheringError{ // nolint
		Type:    "xml-compile-error",
		Message: "error compiling provided xml file",
	}

	XmlPathValueNotFoundError = entities.FactGatheringError{ // nolint
		Type:    "xml-xpath-value-not-found",
		Message: "requested xpath value not found",
	}
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
			gatheringError := XmlCompileError.Wrap(err.Error())
			log.Errorf(gatheringError.Error())
			return entities.NewFactsGatheredListWithError(factsRequests, &gatheringError), nil
		}

		var fact entities.FactsGatheredItem

		value, ok := x.String(root)
		if ok {
			fact = entities.NewFactGatheredWithRequest(factReq, value)
		} else {
			gatheringError := XmlPathValueNotFoundError.Wrap(fmt.Sprintf("requested xpath %s not found", factReq.Argument))
			log.Errorf(gatheringError.Error())
			fact = entities.NewFactGatheredWithError(factReq, &gatheringError)
		}

		facts = append(facts, fact)
	}

	return facts, nil
}
