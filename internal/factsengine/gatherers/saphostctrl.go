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
	saphostCtrlPingParsingRegexp          = `(SUCCESS) \( *(\d+) usec\)`
)

// nolint:gochecknoglobals
var (
	SapHostCtrlCommandError = entities.FactGatheringError{
		Type:    "saphostctrl-cmd-error",
		Message: "error executing saphostctrl command",
	}
	SapHostCtrlUnsupportedFunction = entities.FactGatheringError{
		Type:    "saphostctrl-webmethod-error",
		Message: "requested webmethod not whitelisted",
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
	whitelistedParsers := whitelistedMethodsParsers()

	for _, factReq := range factsRequests {
		var fact entities.Fact
		if _, ok := whitelistedParsers[factReq.Argument]; !ok {
			gatheringError := SapHostCtrlUnsupportedFunction.Wrap(factReq.Argument)
			log.Error(gatheringError)
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)
			facts = append(facts, fact)
			continue
		}
		saphostctlOutput, execErr := g.executor.Exec("/usr/sap/hostctrl/exe/saphostctrl", "-function", factReq.Argument)
		if execErr != nil {
			gatheringError := SapHostCtrlCommandError.Wrap(factReq.Argument)
			log.Error(gatheringError)
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)
			facts = append(facts, fact)
			continue
		}

		parsedOutput, parseErr := whitelistedParsers[factReq.Argument](string(saphostctlOutput))
		if parseErr != nil {
			fact = entities.NewFactGatheredWithError(factReq, parseErr)
			facts = append(facts, fact)
			continue
		}
		fact = entities.NewFactGatheredWithRequest(factReq, parsedOutput)

		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", SapHostCtrlGathererName)
	return facts, nil
}

func whitelistedMethodsParsers() map[string]func(string) (entities.FactValue, *entities.FactGatheringError) {
	whitelist := map[string]func(string) (entities.FactValue, *entities.FactGatheringError){
		"Ping":          parsePing,
		"ListInstances": parseInstance,
	}

	return whitelist
}

func parsePing(commandOutput string) (entities.FactValue, *entities.FactGatheringError) {
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

func parseInstance(commandOutput string) (entities.FactValue, *entities.FactGatheringError) {
	r, _ := regexp.Compile(saphostCtrlListInstancesParsingRegexp)
	lines := strings.Split(commandOutput, "\n")
	instances := []entities.FactValue{}

	for _, line := range lines {
		instance := map[string]entities.FactValue{}
		if r.MatchString(line) {
			fields := r.FindStringSubmatch(line)
			if len(fields) < 6 {
				return nil, SapHostCtrlParseError.Wrap(commandOutput)
			}

			instance["system"] = &entities.FactValueString{Value: fields[1]}
			instance["instance"] = &entities.FactValueString{Value: fields[2]}
			instance["hostname"] = &entities.FactValueString{Value: fields[3]}
			instance["revision"] = &entities.FactValueString{Value: fields[4]}
			instance["patch"] = &entities.FactValueString{Value: fields[5]}
			instance["changelist"] = &entities.FactValueString{Value: fields[6]}

			instances = append(instances, &entities.FactValueMap{Value: instance})
		}
	}

	result := &entities.FactValueList{Value: instances}

	return result, nil
}
