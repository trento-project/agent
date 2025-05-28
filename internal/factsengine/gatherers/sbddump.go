package gatherers

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/core/cluster"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	SBDDumpGathererName = "sbd_dump"
)

var undesiredParenthesesRegexp = regexp.MustCompile(`[()]`)

// nolint:gochecknoglobals
var (
	SBDDevicesLoadingError = entities.FactGatheringError{
		Type:    "sbd-devices-loading-error",
		Message: "error loading the configured sbd devices",
	}

	SBDDumpCommandError = entities.FactGatheringError{
		Type:    "sbd-dump-command-error",
		Message: "error while executing sbd dump",
	}
)

type SBDDumpGatherer struct {
	executor      utils.CommandExecutor
	sbdConfigFile string
}

func NewDefaultSBDDumpGatherer() *SBDDumpGatherer {
	return NewSBDDumpGatherer(utils.Executor{}, cluster.SBDConfigPath)
}

func NewSBDDumpGatherer(executor utils.CommandExecutor, sbdConfigFile string) *SBDDumpGatherer {
	return &SBDDumpGatherer{
		executor:      executor,
		sbdConfigFile: sbdConfigFile,
	}
}

func (gatherer *SBDDumpGatherer) Gather(
	ctx context.Context,
	factsRequests []entities.FactRequest,
) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting %s facts gathering process", SBDDumpGathererName)

	configuredDevices, err := loadDevices(gatherer.sbdConfigFile)

	if err != nil {
		return nil, SBDDevicesLoadingError.Wrap(err.Error())
	}

	for _, factRequest := range factsRequests {
		var fact entities.Fact

		if devicesDumps, err := getSBDDevicesDumps(ctx, gatherer.executor, configuredDevices); err == nil {
			fact = entities.NewFactGatheredWithRequest(factRequest, devicesDumps)
		} else {
			gatheringError := SBDDumpCommandError.Wrap(err.Error())

			fact = entities.NewFactGatheredWithError(factRequest, gatheringError)
		}
		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", SBDDumpGathererName)

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return facts, nil
}

func loadDevices(sbdConfigFile string) ([]string, error) {
	conf, err := cluster.LoadSbdConfig(sbdConfigFile)

	if err != nil {
		return nil, err
	}

	configuredDevices, found := conf["SBD_DEVICE"]

	if !found || configuredDevices == "" {
		return nil, errors.New("unable to load configured devices")
	}

	return strings.Split(configuredDevices, ";"), nil
}

func getSBDDevicesDumps(
	ctx context.Context,
	executor utils.CommandExecutor,
	configuredDevices []string) (*entities.FactValueMap, error) {
	var devicesDumps = make(map[string]entities.FactValue)

	for _, device := range configuredDevices {
		SBDDumpMap, err := getSBDDumpFactValueMap(ctx, executor, device)
		if err != nil {
			log.Errorf("Error getting sbd dump for device: %s", device)

			return nil, err
		}

		devicesDumps[device] = SBDDumpMap
	}

	return &entities.FactValueMap{Value: devicesDumps}, nil
}

func getSBDDumpFactValueMap(
	ctx context.Context,
	executor utils.CommandExecutor,
	device string) (*entities.FactValueMap, error) {
	SBDDump, err := executor.ExecContext(ctx, "/usr/sbin/sbd", "-d", device, "dump")
	if err != nil {
		return nil, fmt.Errorf("Error while dumping information for device %s: %w", device, err)
	}

	SBDDumpMap := utils.FindMatches(`(?m)^(\S+(?: \S+)*)\s*:\s(\S*)$`, SBDDump)

	var deviceDump = make(map[string]entities.FactValue)

	for key, value := range SBDDumpMap {
		key = undesiredParenthesesRegexp.ReplaceAllString(key, "")

		deviceDump[strings.ToLower(key)] = entities.ParseStringToFactValue(fmt.Sprint(value))
	}

	return &entities.FactValueMap{Value: deviceDump}, nil
}
