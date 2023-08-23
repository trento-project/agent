package gatherers

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-envparse"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	PackageVersionGathererName = "package_version"
	invalidVersionCompare      = -2
	packageVersionQueryFormat  = "VERSION=%{VERSION}\nINSTALLTIME=%{INSTALLTIME}\n---\n"
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

type packageVersion struct {
	Version     string
	InstalledOn time.Time
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

		installedVersions, err := executeRpmVersionRetrieveCommand(g.executor, packageName)
		if err != nil {
			fact = entities.NewFactGatheredWithError(factReq, err)
			facts = append(facts, fact)
			continue
		}

		if requestedVersion != "" {
			comparisonResult, err := executeZypperVersionCmpCommand(g.executor, installedVersions[0].Version, requestedVersion)
			if err != nil {
				fact = entities.NewFactGatheredWithError(factReq, err)
				facts = append(facts, fact)
				continue
			}
			fact = entities.NewFactGatheredWithRequest(factReq, &entities.FactValueInt{Value: comparisonResult})
			facts = append(facts, fact)
			continue
		}

		fact = entities.NewFactGatheredWithRequest(factReq, installedVersionsToFactValueList(installedVersions))
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
	outputParts := strings.Split(versionCmpResult, "\n")
	comparisonResult := outputParts[len(outputParts)-1]

	result, err := strconv.ParseInt(comparisonResult, 10, 32)
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
) ([]packageVersion, *entities.FactGatheringError) {
	rpmOutputBytes, err := executor.Exec("/usr/bin/rpm", "-q", "--qf", packageVersionQueryFormat, packageName)

	rpmOutput := string(rpmOutputBytes)

	if err != nil {
		gatheringError := PackageVersionRpmCommandError.Wrap(rpmOutput + err.Error())
		log.Error(gatheringError)
		return nil, gatheringError
	}

	installedVersions := []packageVersion{}

	for _, detectedVersionLine := range strings.Split(rpmOutput, "\n---\n") {
		if detectedVersionLine == "" {
			continue
		}

		packageVersionInfo, err := envparse.Parse(strings.NewReader(detectedVersionLine))

		if err != nil {
			parsingError := fmt.Sprintf("Unable to parse rpm output: %s, output:%s", err.Error(), rpmOutput)
			return nil, PackageVersionRpmCommandError.Wrap(parsingError)
		}

		detectedPackageInstallationTime := packageVersionInfo["INSTALLTIME"]
		detectedPackageVersion := packageVersionInfo["VERSION"]

		timestamp, err := strconv.ParseInt(detectedPackageInstallationTime, 10, 64)
		if err != nil {
			invalidDateError := fmt.Sprintf("Unable to parse package installation timestamp to an integer: %s", err.Error())
			return nil, PackageVersionRpmCommandError.Wrap(invalidDateError)
		}

		installedVersions = append(installedVersions, packageVersion{
			Version:     detectedPackageVersion,
			InstalledOn: time.Unix(timestamp, 0).UTC(),
		})
	}

	sort.Slice(installedVersions, func(i, j int) bool {
		return installedVersions[i].InstalledOn.After(installedVersions[j].InstalledOn)
	})

	return installedVersions, nil
}

func installedVersionsToFactValueList(installedVersions []packageVersion) *entities.FactValueList {
	installedVersionsValue := []entities.FactValue{}
	for _, installedVersion := range installedVersions {
		installedVersionsValue = append(installedVersionsValue, &entities.FactValueMap{Value: map[string]entities.FactValue{
			"version": entities.ParseStringToFactValue(installedVersion.Version),
		}})
	}

	return &entities.FactValueList{Value: installedVersionsValue}
}
