package gatherers

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	SBDDumpGathererName = "sbd_dump"
)

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
	executor    utils.CommandExecutor
	sbdGatherer *SBDGatherer
}

func NewDefaultSBDDumpGatherer() *SBDDumpGatherer {
	return NewSBDDumpGatherer(utils.Executor{}, NewDefaultSBDGatherer())
}

func NewSBDDumpGatherer(executor utils.CommandExecutor, sbdGatherer *SBDGatherer) *SBDDumpGatherer {
	return &SBDDumpGatherer{
		executor:    executor,
		sbdGatherer: sbdGatherer,
	}
}

func (gatherer *SBDDumpGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting %s facts gathering process", SBDDumpGathererName)

	configuredDevices, err := loadDevices(gatherer.sbdGatherer)

	if err != nil {
		return facts, SBDDevicesLoadingError.Wrap(err.Error())
	}

	for _, factRequest := range factsRequests {
		var fact entities.Fact

		if devicesDumps, err := getSBDDevicesDumps(gatherer.executor, configuredDevices); err == nil {
			fact = entities.NewFactGatheredWithRequest(factRequest, &entities.FactValueMap{Value: devicesDumps})
		} else {
			gatheringError := SBDDumpCommandError.Wrap(err.Error())

			fact = entities.NewFactGatheredWithError(factRequest, gatheringError)
		}
		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", SBDDumpGathererName)
	return facts, nil
}

func loadDevices(gatherer *SBDGatherer) ([]string, error) {
	sbdDevicesRequest := []entities.FactRequest{
		{
			Name:     "configured_devices",
			Gatherer: "sbd_config",
			Argument: "SBD_DEVICE",
		},
	}

	sbdDevicesFacts, err := gatherer.Gather(sbdDevicesRequest)

	if err != nil {
		return []string{}, err
	}

	sbdDevicesFact := sbdDevicesFacts[0]

	if sbdDevicesFact.Error != nil {
		return []string{}, sbdDevicesFact.Error
	}

	deviceAsStringValue, ok := sbdDevicesFact.Value.(*entities.FactValueString)

	if !ok {
		return []string{}, fmt.Errorf("Unable to determine the device name: %s", sbdDevicesFact.Value)
	}

	return strings.Split(deviceAsStringValue.Value, ";"), nil
}

func getSBDDevicesDumps(
	executor utils.CommandExecutor,
	configuredDevices []string) (map[string]entities.FactValue, error) {
	var devicesDumps = make(map[string]entities.FactValue)

	for _, device := range configuredDevices {
		SBDDumpMap, err := getSBDDumpFactValueMap(executor, device)
		if err != nil {
			log.Errorf("Error getting sbd dump for device: %s", device)

			return nil, err
		}

		devicesDumps[device] = SBDDumpMap
	}

	return devicesDumps, nil
}

func getSBDDumpFactValueMap(executor utils.CommandExecutor, device string) (*entities.FactValueMap, error) {
	SBDDump, err := executor.Exec("sbd", "-d", device, "dump")
	if err != nil {
		return nil, err
	}

	SBDDumpMap := utils.FindMatches(`(?m)^(\S+(?: \S+)*)\s*:\s(\S*)$`, SBDDump)

	var deviceDump = make(map[string]entities.FactValue)

	for key, value := range SBDDumpMap {
		valueAsString, ok := value.(string)

		if !ok {
			err := fmt.Errorf(`Unable to parse sbd dump entry "%s" as string: %s`, key, value)
			log.Error(err)
			return nil, err
		}

		key = regexp.MustCompile(`[()]`).ReplaceAllString(key, "")

		deviceDump[strings.ToLower(key)] = entities.ParseStringToFactValue(valueAsString)
	}

	return &entities.FactValueMap{Value: deviceDump}, nil
}
