package gatherers

import (
	log "github.com/sirupsen/logrus"
)

const (
	SapHostCtrlGathererName = "saphostctrl"
)

type SapHostCtrlGatherer struct {
	executor CommandExecutor
}

func NewDefaultSapHostCtrlGatherer() *SapHostCtrlGatherer {
	return NewSapHostCtrlGatherer(Executor{})
}

func NewSapHostCtrlGatherer(executor CommandExecutor) *SapHostCtrlGatherer {
	return &SapHostCtrlGatherer{
		executor: executor,
	}
}

func (g *SapHostCtrlGatherer) Gather(factsRequests []FactRequest) ([]Fact, error) {
	facts := []Fact{}
	log.Infof("Starting saphostctrl facts gathering process")

	for _, factReq := range factsRequests {
		// TODO: The instance number "00" could be an additional argument in the future
		version, err := g.executor.Exec(
			"/usr/sap/hostctrl/exe/saphostctrl", "-nr", "00", "-function", factReq.Argument)
		if err != nil {
			// TODO: Decide together with Wanda how to deal with errors. `err` field in the fact result?
			log.Errorf("Error running saphostctrl: %s", err)
		}
		fact := NewFactWithRequest(factReq, string(version))
		facts = append(facts, fact)
	}

	log.Infof("Requested saphostctrl facts gathered")
	return facts, nil
}
