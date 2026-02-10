package gatherers

import (
	"context"
	"log/slog"
	"regexp"
	"strings"

	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	SapHostCtrlGathererName = "saphostctrl"
)

// nolint:gochecknoglobals
var (
	saphostCtrlListInstancesParsingRegexp = regexp.MustCompile(`^\s+Inst Info\s*` +
		`:\s*([^-]+?)\s*-\s*(\d+)\s*-\s*([^,]+?)` +
		`\s*-\s*(\d+),\s*patch\s*(\d+),\s*changelist\s*(\d+)$`)
	saphostCtrlPingParsingRegexp = regexp.MustCompile(`(SUCCESS|FAILED) \( *(\d+) usec\)`)
)

// nolint:gochecknoglobals
var whitelistedWebmethods = map[string]func(string) (entities.FactValue, *entities.FactGatheringError){
	"Ping":          parsePing,
	"ListInstances": parseInstances,
}

// nolint:gochecknoglobals
var (
	SapHostCtrlCommandError = entities.FactGatheringError{
		Type:    "saphostctrl-cmd-error",
		Message: "error executing saphostctrl command",
	}

	SapHostCtrlUnsupportedFunction = entities.FactGatheringError{
		Type:    "saphostctrl-webmethod-error",
		Message: "requested webmethod not supported",
	}

	SapHostCtrlParseError = entities.FactGatheringError{
		Type:    "saphostctrl-parse-error",
		Message: "error while parsing saphostctrl output",
	}

	SapHostCtrlMissingArgument = entities.FactGatheringError{
		Type:    "saphostctrl-missing-argument",
		Message: "missing required argument",
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

func (g *SapHostCtrlGatherer) Gather(
	ctx context.Context,
	factsRequests []entities.FactRequest,
) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	slog.Info("Starting saphostctrl facts gathering process")

	for _, factReq := range factsRequests {
		var fact entities.Fact
		if len(factReq.Argument) == 0 {
			slog.Error(SapHostCtrlMissingArgument.Message)
			fact = entities.NewFactGatheredWithError(factReq, &SapHostCtrlMissingArgument)
		} else if factValue, err := handleWebmethod(ctx, g.executor, factReq.Argument); err != nil {
			slog.Error(err.Error())
			fact = entities.NewFactGatheredWithError(factReq, err)
		} else {
			fact = entities.NewFactGatheredWithRequest(factReq, factValue)
		}

		facts = append(facts, fact)
	}

	slog.Info("Requested facts gathered", "gatherer", SapHostCtrlGathererName)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return facts, nil
}

func handleWebmethod(
	ctx context.Context,
	executor utils.CommandExecutor,
	webMethod string,
) (entities.FactValue, *entities.FactGatheringError) {
	webMethodHandler, ok := whitelistedWebmethods[webMethod]

	if !ok {
		gatheringError := SapHostCtrlUnsupportedFunction.Wrap(webMethod)
		slog.Error(gatheringError.Error())
		return nil, gatheringError
	}

	saphostctlOutput, commandError := executeSapHostCtrlCommand(ctx, executor, webMethod)
	if commandError != nil {
		slog.Error(commandError.Error())
		return nil, commandError
	}

	return webMethodHandler(saphostctlOutput)
}

func executeSapHostCtrlCommand(
	ctx context.Context,
	executor utils.CommandExecutor,
	command string,
) (string, *entities.FactGatheringError) {
	saphostctlOutput, err := executor.OutputContext(ctx, "/usr/sap/hostctrl/exe/saphostctrl", "-function", command)
	if err != nil {
		gatheringError := SapHostCtrlCommandError.Wrap(err.Error())
		slog.Error(gatheringError.Error())
		return "", gatheringError
	}

	return string(saphostctlOutput), nil
}

func parsePing(commandOutput string) (entities.FactValue, *entities.FactGatheringError) {
	pingData := map[string]entities.FactValue{}

	matches := saphostCtrlPingParsingRegexp.FindStringSubmatch(commandOutput)
	if len(matches) < 2 {
		return nil, SapHostCtrlParseError.Wrap(commandOutput)
	}

	pingData["status"] = &entities.FactValueString{Value: matches[1]}
	pingData["elapsed"] = entities.ParseStringToFactValue(matches[2])

	result := &entities.FactValueMap{Value: pingData}

	return result, nil
}

func parseInstances(commandOutput string) (entities.FactValue, *entities.FactGatheringError) {
	lines := strings.Split(commandOutput, "\n")
	instances := []entities.FactValue{}

	for _, line := range lines {
		instance := map[string]entities.FactValue{}
		if saphostCtrlListInstancesParsingRegexp.MatchString(line) {
			fields := saphostCtrlListInstancesParsingRegexp.FindStringSubmatch(line)
			if len(fields) < 6 {
				return nil, SapHostCtrlParseError.Wrap(commandOutput)
			}

			instance["system"] = &entities.FactValueString{Value: fields[1]}
			instance["instance"] = &entities.FactValueString{Value: fields[2]}
			instance["hostname"] = &entities.FactValueString{Value: fields[3]}
			instance["sapkernel"] = entities.ParseStringToFactValue(fields[4])
			instance["patch"] = entities.ParseStringToFactValue(fields[5])
			instance["changelist"] = entities.ParseStringToFactValue(fields[6])

			instances = append(instances, &entities.FactValueMap{Value: instance})
		}
	}

	result := &entities.FactValueList{Value: instances}

	return result, nil
}
