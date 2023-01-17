package gatherers

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	PackageVersionGathererName = "package_version"
	invalidVersionCompare      = -2
)

// nolint:gochecknoglobals
var (
	PackageVersionRpmCommandError = entities.FactGatheringError{
		Type:    "package-version-rpm-cmd-error",
		Message: "error while fetching package version",
	}

	PackageVersionZypperCommandError = entities.FactGatheringError{
		Type:    "package-version-zypper-cmd-error",
		Message: "error while executing zypper",
	}

	PackageVersionMissingArgument = entities.FactGatheringError{
		Type:    "package-version-missing-argument",
		Message: "missing required argument",
	}
)

type PackageVersionGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultPackageVersionGatherer() *PackageVersionGatherer {
	return NewPackageVersionGatherer(utils.Executor{})
}

func NewPackageVersionGatherer(executor utils.CommandExecutor) *PackageVersionGatherer {
	return &PackageVersionGatherer{
		executor: executor,
	}
}

func (g *PackageVersionGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting %s facts gathering process", PackageVersionGathererName)

	for _, factReq := range factsRequests {
		var fact entities.Fact
		if len(factReq.Argument) == 0 {
			log.Error(PackageVersionMissingArgument.Message)
			fact = entities.NewFactGatheredWithError(factReq, &PackageVersionMissingArgument)
			facts = append(facts, fact)
			continue
		}

		packageName := factReq.Argument
		requestedVersion := ""
		if strings.Contains(factReq.Argument, ",") {
			arguments := strings.SplitN(factReq.Argument, ",", 2)
			packageName = arguments[0]
			requestedVersion = arguments[1]
		}

		installedVersion, err := executeRpmVersionRetrieveCommand(g.executor, packageName)
		if err != nil {
			fact = entities.NewFactGatheredWithError(factReq, err)
			facts = append(facts, fact)
			continue
		}

		if requestedVersion != "" {
			comparisonResult, err := executeZypperVersionCmpCommand(g.executor, installedVersion, requestedVersion)
			if err != nil {
				fact = entities.NewFactGatheredWithError(factReq, err)
				facts = append(facts, fact)
				continue
			}
			fact = entities.NewFactGatheredWithRequest(factReq, &entities.FactValueInt{Value: comparisonResult})
			facts = append(facts, fact)
			continue
		}

		fact = entities.NewFactGatheredWithRequest(factReq, &entities.FactValueString{Value: installedVersion})
		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", PackageVersionGathererName)
	return facts, nil
}

func executeZypperVersionCmpCommand(
	executor utils.CommandExecutor,
	installedVersion string,
	comparedVersion string,
) (int, *entities.FactGatheringError) {
	zypperOutput, err := executor.Exec("/usr/bin/zypper", "--terse", "versioncmp", comparedVersion, installedVersion)
	if err != nil {
		gatheringError := PackageVersionZypperCommandError.Wrap(err.Error())
		log.Error(gatheringError)
		return invalidVersionCompare, gatheringError
	}

	versionCmpResult := strings.TrimRight(string(zypperOutput), "\n")
	result, err := strconv.ParseInt(versionCmpResult, 10, 32)
	if err != nil {
		gatheringError := PackageVersionRpmCommandError.Wrap(err.Error())
		log.Error(gatheringError)
		return invalidVersionCompare, gatheringError
	}

	return int(result), nil
}

func executeRpmVersionRetrieveCommand(
	executor utils.CommandExecutor,
	packageName string,
) (string, *entities.FactGatheringError) {
	rpmOutput, err := executor.Exec("/usr/bin/rpm", "-q", "--qf", "%{VERSION}", packageName)
	if err != nil {
		gatheringError := PackageVersionRpmCommandError.Wrap(string(rpmOutput) + err.Error())
		log.Error(gatheringError)
		return "", gatheringError
	}

	return string(rpmOutput), nil
}
