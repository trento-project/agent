package gatherers

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/trento-project/agent/internal/core/sapsystem"
	"github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi"
	"github.com/trento-project/agent/internal/factsengine/factscache"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	SapControlGathererName  = "sapcontrol"
	SapControlGathererCache = "sapcontrol"
)

// nolint:gochecknoglobals
var whitelistedSapControlArguments = map[string]func(context.Context, sapcontrolapi.WebService) (interface{}, error){
	"GetProcessList":        mapGetProcessList,
	"GetSystemInstanceList": mapGetSystemInstanceList,
	"GetVersionInfo":        mapGetVersionInfo,
	"HACheckConfig":         mapHACheckConfig,
	"HAGetFailoverConfig":   mapHAGetFailoverConfig,
}

// nolint:gochecknoglobals
var (
	SapcontrolFileSystemError = entities.FactGatheringError{
		Type:    "sapcontrol-file-system-error",
		Message: "error in the SAP file system",
	}

	SapcontrolArgumentUnsupported = entities.FactGatheringError{
		Type:    "sapcontrol-unsupported-argument",
		Message: "the requested argument is not currently supported",
	}

	SapcontrolMissingArgument = entities.FactGatheringError{
		Type:    "sapcontrol-missing-argument",
		Message: "missing required argument",
	}

	SapcontrolWebmethodError = entities.FactGatheringError{
		Type:    "sapcontrol-webmethod-error",
		Message: "error executing sapcontrol webmethod",
	}

	SapcontrolDecodingError = entities.FactGatheringError{
		Type:    "sapcontrol-decoding-error",
		Message: "error decoding sapcontrol output",
	}

	versionInfoPatternCompiled = regexp.MustCompile("^(\\d+), patch (\\d+), changelist (\\d+), " +
		"RKS compatibility level (\\d+), (.*), (.*)$")
)

type versionInfo struct {
	Filename              string `json:"filename,omitempty"`
	SapKernel             string `json:"sap_kernel,omitempty"`
	Patch                 string `json:"patch,omitempty"`
	ChangeList            string `json:"changelist,omitempty"`
	RKSCompatibilityLevel string `json:"rks_compatibility_level,omitempty"`
	Build                 string `json:"build,omitempty"`
	Architecture          string `json:"architecture,omitempty"`
	Time                  string `json:"time,omitempty"`
}

type failoverConfig struct {
	HAActive              bool     `json:"ha_active"`
	HAProductVersion      string   `json:"ha_product_version"`
	HASAPInterfaceVersion string   `json:"ha_sap_interface_version"`
	HADocumentation       string   `json:"ha_documentation"`
	HAActiveNode          string   `json:"ha_active_nodes"`
	HANodes               []string `json:"ha_nodes"`
}

type SapControlMap map[string][]SapControlInstance

type SapControlInstance struct {
	Name       string      `json:"name"`
	InstanceNr string      `json:"instance_nr"`
	Output     interface{} `json:"output"`
}

type SapControlGatherer struct {
	webService sapcontrolapi.WebServiceConnector
	fs         afero.Fs
	cache      *factscache.FactsCache
}

func NewDefaultSapControlGatherer() *SapControlGatherer {
	webService := sapcontrolapi.WebServiceUnix{}
	fs := afero.NewOsFs()
	return NewSapControlGatherer(webService, fs, nil)
}

func NewSapControlGatherer(
	webService sapcontrolapi.WebServiceConnector,
	fs afero.Fs,
	cache *factscache.FactsCache) *SapControlGatherer {

	return &SapControlGatherer{
		webService: webService,
		fs:         fs,
		cache:      cache,
	}
}

func (s *SapControlGatherer) SetCache(cache *factscache.FactsCache) {
	s.cache = cache
}

func memoizeSapcontrol(args ...interface{}) (interface{}, error) {
	ctx, ok := args[0].(context.Context)
	if !ok {
		return nil, ImplementationError.Wrap("error using memoizeSapcontrol. Context must be 1st argument")
	}

	webService, ok := args[1].(sapcontrolapi.WebServiceConnector)
	if !ok {
		return nil, ImplementationError.Wrap("error using memoizeSapcontrol. WebServiceConnector must be 2nd argument")
	}

	instanceNumber, ok := args[2].(string)
	if !ok {
		return nil, ImplementationError.Wrap("error using memoizeSapcontrol. string must be 3rd argument")
	}

	webmethod, ok := args[3].(func(ctx context.Context, conn sapcontrolapi.WebService) (interface{}, error))
	if !ok {
		return nil, ImplementationError.Wrap("error using memoizeSapcontrol. webmethod func must be 4th argument")
	}

	conn := webService.New(instanceNumber)
	return webmethod(ctx, conn)
}

func (s *SapControlGatherer) Gather(_ context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	ctx := context.Background()

	log.Infof("Starting %s facts gathering process", SapControlGathererName)
	facts := []entities.Fact{}

	foundSystems, err := initSystemsMap(s.fs)
	if err != nil {
		return nil, SapcontrolFileSystemError.Wrap(err.Error())
	}

	for _, factReq := range factsRequests {
		if len(factReq.Argument) == 0 {
			log.Error(SapcontrolMissingArgument.Error())
			facts = append(facts, entities.NewFactGatheredWithError(factReq, &SapcontrolMissingArgument))
			continue
		}

		webmethod, ok := whitelistedSapControlArguments[factReq.Argument]

		if !ok {
			gatheringError := SapcontrolArgumentUnsupported.Wrap(factReq.Argument)
			log.Error(gatheringError)
			facts = append(facts, entities.NewFactGatheredWithError(factReq, gatheringError))
			continue
		}

		sapControlMap := make(SapControlMap)
		for sid, instances := range foundSystems {
			sapControlInstance := []SapControlInstance{}
			for _, instanceData := range instances {
				instanceName, instanceNumber := instanceData[0], instanceData[1]
				cacheEntry := fmt.Sprintf("%s:%s:%s:%s", SapControlGathererCache, factReq.Argument, sid, instanceNumber)
				output, err := factscache.GetOrUpdate(
					s.cache,
					cacheEntry,
					memoizeSapcontrol,
					ctx,
					s.webService,
					instanceNumber,
					webmethod,
				)

				if err != nil {
					log.Error(SapcontrolWebmethodError.
						Wrap(fmt.Sprintf("argument %s for %s/%s", factReq.Argument, sid, instanceName)).
						Wrap(err.Error()))
					continue
				}
				sapControlInstance = append(sapControlInstance, SapControlInstance{
					Name:       instanceName,
					InstanceNr: instanceNumber,
					Output:     output,
				})
				sapControlMap[sid] = sapControlInstance
			}
		}

		var fact entities.Fact

		if factValue, err := outputToFactValue(sapControlMap); err != nil {
			gatheringError := SapcontrolDecodingError.
				Wrap(fmt.Sprintf("argument: %s", factReq.Argument)).
				Wrap(err.Error())
			log.Error(gatheringError)
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)
		} else {
			fact = entities.NewFactGatheredWithRequest(factReq, factValue)
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", SapControlGathererName)

	return facts, nil
}

func initSystemsMap(fs afero.Fs) (map[string][][]string, error) {
	foundSystems := make(map[string][][]string)
	systems, err := sapsystem.FindSystems(fs)
	if err != nil {
		return nil, err
	}

	for _, system := range systems {
		sid := filepath.Base(system)
		instances, err := sapsystem.FindInstances(fs, system)
		if err != nil {
			return nil, err
		}

		foundSystems[sid] = instances
	}

	return foundSystems, err
}

func mapGetProcessList(ctx context.Context, conn sapcontrolapi.WebService) (interface{}, error) {
	output, err := conn.GetProcessList(ctx)
	if err != nil {
		return nil, err
	}

	return output.Processes, nil
}

func mapGetSystemInstanceList(ctx context.Context, conn sapcontrolapi.WebService) (interface{}, error) {
	output, err := conn.GetSystemInstanceList(ctx)
	if err != nil {
		return nil, err
	}

	return output.Instances, nil
}

func mapGetVersionInfo(ctx context.Context, conn sapcontrolapi.WebService) (interface{}, error) {
	output, err := conn.GetVersionInfo(ctx)
	if err != nil {
		return nil, err
	}

	versions := []versionInfo{}

	for _, version := range output.InstanceVersions {
		fields := versionInfoPatternCompiled.FindStringSubmatch(version.VersionInfo)
		if len(fields) != 7 {
			return nil, fmt.Errorf("incorrect number of fields in line %s", version.VersionInfo)
		}

		versions = append(versions, versionInfo{
			Filename:              version.Filename,
			SapKernel:             fields[1],
			Patch:                 fields[2],
			ChangeList:            fields[3],
			RKSCompatibilityLevel: fields[4],
			Build:                 fields[5],
			Architecture:          fields[6],
			Time:                  version.Time,
		})
	}

	return versions, nil
}

func mapHACheckConfig(ctx context.Context, conn sapcontrolapi.WebService) (interface{}, error) {
	output, err := conn.HACheckConfig(ctx)
	if err != nil {
		return nil, err
	}

	return output.Checks, nil
}

func mapHAGetFailoverConfig(ctx context.Context, conn sapcontrolapi.WebService) (interface{}, error) {
	output, err := conn.HAGetFailoverConfig(ctx)
	if err != nil {
		return nil, err
	}

	haNodes := []string{}
	if output.HANodes != nil {
		haNodes = *output.HANodes
	}

	config := failoverConfig{
		HAActive:              output.HAActive,
		HAProductVersion:      output.HAProductVersion,
		HASAPInterfaceVersion: output.HASAPInterfaceVersion,
		HADocumentation:       output.HADocumentation,
		HAActiveNode:          output.HAActiveNode,
		HANodes:               haNodes,
	}

	return config, nil
}

func outputToFactValue(output interface{}) (*entities.FactValueMap, error) {
	marshalled, err := json.Marshal(&output)
	if err != nil {
		return nil, err
	}

	var unmarshalled map[string]interface{}
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return nil, err
	}

	// Trick to keep the SIDs as capital letter
	result := &entities.FactValueMap{Value: make(map[string]entities.FactValue)}
	for key, value := range unmarshalled {
		factValue, err := entities.NewFactValue(value, entities.WithSnakeCaseKeys())
		if err != nil {
			return nil, err
		}
		result.Value[key] = factValue
	}

	return result, nil
}
