package gatherers

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	SapControlGathererName = "sapcontrol"
)

// nolint:gochecknoglobals
var (
	sapControlGetProcessListRegexp        = regexp.MustCompile(`^([^,]+),\s+(.+),\s+(.+),\s+(.+),\s+(.+),\s+(.+),\s+(\d+)`)
	sapControlGetSystemInstanceListRegexp = regexp.MustCompile(`^([^,]+),\s+(\d+),\s+(\d+),\s+(\d+),\s+(.+),\s+(.+),\s+(.+)`)
)

// nolint:gochecknoglobals
var whitelistedSapControlWebmethods = map[string]func(string) (entities.FactValue, *entities.FactGatheringError){
	"CheckHostAgent":        parsecheckHostAgent,
	"GetProcessList":        parseGetProcessList,
	"GetSystemInstanceList": parseGetSystemInstanceList,
}

// nolint:gochecknoglobals
var (
	SapControlCommandError = entities.FactGatheringError{
		Type:    "sapcontrol-cmd-error",
		Message: "error executing sapcontrol command",
	}

	SapControlMissingArgument = entities.FactGatheringError{
		Type:    "sapcontrol-missing-argument",
		Message: "missing required argument",
	}

	SapControlUnsupportedFunction = entities.FactGatheringError{
		Type:    "sapcontrol-webmethod-error",
		Message: "requested webmethod not supported",
	}

	SapControlParseError = entities.FactGatheringError{
		Type:    "sapcontrol-parse-error",
		Message: "error while parsing sapcontrol output",
	}
)

type SapControlGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultSapControlGatherer() *SapControlGatherer {
	return NewSapControlGatherer(utils.Executor{})
}

func NewSapControlGatherer(executor utils.CommandExecutor) *SapControlGatherer {
	return &SapControlGatherer{
		executor: executor,
	}
}

func (s *SapControlGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting sapcontrol facts gathering process")

	for _, factReq := range factsRequests {
		var fact entities.Fact
		if len(factReq.Argument) == 0 {
			log.Error(SapControlMissingArgument.Message)
			fact = entities.NewFactGatheredWithError(factReq, &SapControlMissingArgument)
		} else if factValue, err := handleSapControlWebMethod(s.executor, factReq.Argument); err != nil {
			fact = entities.NewFactGatheredWithError(factReq, err)
		} else {
			fact = entities.NewFactGatheredWithRequest(factReq, factValue)
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", SapControlGathererName)
	return facts, nil
}

func handleSapControlWebMethod(
	executor utils.CommandExecutor,
	webMethod string,
) (entities.FactValue, *entities.FactGatheringError) {
	webMethodHandler, ok := whitelistedSapControlWebmethods[webMethod]

	if !ok {
		gatheringError := SapControlUnsupportedFunction.Wrap(webMethod)
		log.Error(gatheringError)
		return nil, gatheringError
	}

	sapcontrol, commandError := executeSapControlCommand(executor, webMethod)
	if commandError != nil {
		return nil, commandError
	}

	return webMethodHandler(sapcontrol)
}

func executeSapControlCommand(executor utils.CommandExecutor, command string) (string, *entities.FactGatheringError) {
	sapControlOutput, err := executor.Exec("/usr/sap/hostctrl/exe/sapcontrol", "-nr", "00", "-function", command)
	if err != nil {
		gatheringError := SapControlCommandError.Wrap(err.Error())
		log.Error(gatheringError)
		return "", gatheringError
	}

	return string(sapControlOutput), nil
}

func parsecheckHostAgent(sapControlOutput string) (entities.FactValue, *entities.FactGatheringError) {
	if sapControlOutput == "" {
		gatheringError := SapControlParseError.Wrap("empty output")
		log.Error(gatheringError)
		return nil, gatheringError
	} else {
		return &entities.FactValueBool{Value: (sapControlOutput == "SAPHostAgent Installed\n")}, nil
	}
}

func parseGetProcessList(sapControlOutput string) (entities.FactValue, *entities.FactGatheringError) {
	lines := strings.Split(sapControlOutput, "\n")

	processes := []entities.FactValue{}

	for _, line := range lines {
		if matches := sapControlGetProcessListRegexp.FindStringSubmatch(line); len(matches) > 0 {
			process := map[string]entities.FactValue{}
			process["name"] = &entities.FactValueString{Value: matches[1]}
			process["description"] = &entities.FactValueString{Value: matches[2]}
			process["dispstatus"] = &entities.FactValueString{Value: matches[3]}
			process["textstatus"] = &entities.FactValueString{Value: matches[4]}
			process["starttime"] = &entities.FactValueString{Value: matches[5]}
			process["elapsedtime"] = &entities.FactValueString{Value: matches[6]}
			process["pid"] = entities.ParseStringToFactValue(matches[7])
			processes = append(processes, &entities.FactValueMap{Value: process})
		}
	}

	result := &entities.FactValueList{Value: processes}

	return result, nil
}

func parseGetSystemInstanceList(sapControlOutput string) (entities.FactValue, *entities.FactGatheringError) {
	lines := strings.Split(sapControlOutput, "\n")

	instances := []entities.FactValue{}

	for _, line := range lines {
		if matches := sapControlGetSystemInstanceListRegexp.FindStringSubmatch(line); len(matches) > 0 {
			instance := map[string]entities.FactValue{}

			instance["hostname"] = &entities.FactValueString{Value: matches[1]}
			instance["instanceNr"] = &entities.FactValueString{Value: matches[2]}
			instance["httpPort"] = entities.ParseStringToFactValue(matches[3])
			instance["httpsPort"] = entities.ParseStringToFactValue(matches[4])
			instance["startPriority"] = &entities.FactValueString{Value: matches[5]}
			instance["features"] = &entities.FactValueString{Value: matches[6]}
			instance["dispstatus"] = &entities.FactValueString{Value: matches[7]}

			instances = append(instances, &entities.FactValueMap{Value: instance})
		}
	}

	result := &entities.FactValueList{Value: instances}

	return result, nil
}
