package gatherers

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	SapHostCtrlGathererName               = "saphostctrl"
	saphostCtrlListInstancesParsingRegexp = `^\s+Inst Info\s*:\s*([^-]+?)\s*-\s*(\d+)\s*-\s*([^,]+?)\s*-\s*(\d+),\s*patch\s*(\d+),\s*changelist\s*(\d+)$`
	saphostCtrlPingParsingRegexp          = `(SUCCESS|FAILURE) \( *(\d+) usec\)`
)

// nolint:gochecknoglobals
var (
	SapHostCtrlCommandError = entities.FactGatheringError{
		Type:    "saphostctrl-cmd-error",
		Message: "error executing saphostctrl command",
	}
	SapHostCtrlUnsupportedFunction = entities.FactGatheringError{
		Type:    "saphostctrl-func-error",
		Message: "unsupported saphostctrl function",
	}
	SapHostCtrlParseError = entities.FactGatheringError{
		Type:    "saphostctrl-parse-error",
		Message: "error while parsing saphostctrl output",
	}
)

type SapHostCtrlGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultSapHostCtrlGatherer() *SapHostCtrlGatherer {
	return NewSapHostCtrlGatherer(utils.Executor{})
}

func NewSapHostCtrlGatherer(executor utils.CommandExecutor) *SapHostCtrlGatherer {
	return &SapHostCtrlGatherer{
		executor: executor,
	}
}

func (g *SapHostCtrlGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting saphostctrl facts gathering process")

	for _, factReq := range factsRequests {
		var fact entities.Fact

		if !whitelisted()[factReq.Argument] {
			gatheringError := SapHostCtrlUnsupportedFunction.Wrap(factReq.Argument)
			log.Error(gatheringError)
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)
		} else {
			saphostctlOutput, err := g.executor.Exec("/usr/sap/hostctrl/exe/saphostctrl", "-function", factReq.Argument)
			if err != nil {
				gatheringError := SapHostCtrlCommandError.Wrap(factReq.Argument)
				log.Error(gatheringError)
				fact = entities.NewFactGatheredWithError(factReq, gatheringError)
				facts = append(facts, fact)
				continue
			}
			switch factReq.Argument {
			case "Ping":
				parsedOutput, err := parsePing(string(saphostctlOutput))
				if err != nil {
					fact = entities.NewFactGatheredWithError(factReq, err)
					facts = append(facts, fact)
					continue
				}
				fact = entities.NewFactGatheredWithRequest(factReq, parsedOutput)

			case "ListInstances":
				parsedOutput, err := parseInstances(string(saphostctlOutput))
				if err != nil {
					fact = entities.NewFactGatheredWithError(factReq, err)
					facts = append(facts, fact)
					continue
				}
				fact = entities.NewFactGatheredWithRequest(factReq, parsedOutput)
			}
		}
		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", SapHostCtrlGathererName)
	return facts, nil
}

func whitelisted() map[string]bool {
	whitelist := map[string]bool{"Ping": true, "ListInstances": true}

	return whitelist
}

func parsePing(commandOutput string) (*entities.FactValueMap, *entities.FactGatheringError) {
	re := regexp.MustCompile(saphostCtrlPingParsingRegexp)
	pingData := map[string]entities.FactValue{}

	matches := re.FindStringSubmatch(commandOutput)
	if len(matches) < 2 {
		return nil, SapHostCtrlParseError.Wrap(commandOutput)
	}

	pingData["status"] = &entities.FactValueString{Value: matches[1]}
	pingData["elapsed"] = &entities.FactValueString{Value: matches[2]}

	result := &entities.FactValueMap{Value: pingData}

	return result, nil
}

func parseInstances(commandOutput string) (*entities.FactValueMap, *entities.FactGatheringError) {
	r, _ := regexp.Compile(saphostCtrlListInstancesParsingRegexp)
	lines := strings.Split(commandOutput, "\n")
	instances := map[string]entities.FactValue{}

	for _, line := range lines {
		if r.MatchString(line) {
			fields := r.FindStringSubmatch(line)
			if len(fields) < 6 {
				return nil, SapHostCtrlParseError.Wrap(commandOutput)
			}

			instances["system"] = &entities.FactValueString{Value: fields[1]}
			instances["instance"] = &entities.FactValueString{Value: fields[2]}
			instances["hostname"] = &entities.FactValueString{Value: fields[3]}
			instances["revision"] = &entities.FactValueString{Value: fields[4]}
			instances["patch"] = &entities.FactValueString{Value: fields[5]}
			instances["changelist"] = &entities.FactValueString{Value: fields[6]}
		}
	}

	result := &entities.FactValueMap{Value: instances}

	return result, nil
}
