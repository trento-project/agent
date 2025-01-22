package gatherers

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/core/cluster/cib"
	"github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi"
	"github.com/trento-project/agent/internal/factsengine/factscache"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	AscsErsClusterGathererName = "ascsers_cluster"
	Ensa1                      = "ensa1"
	Ensa2                      = "ensa2"
	EnsaUnknown                = "unknown"
)

// nolint:gochecknoglobals
var (
	AscsErsClusterDecodingError = entities.FactGatheringError{
		Type:    "ascsers-cluster-decoding-error",
		Message: "error decoding cibadmin output",
	}

	AscsErsClusterCibError = entities.FactGatheringError{
		Type:    "ascsers-cluster-cib-error",
		Message: "error parsing cibadmin output",
	}
)

type AscsErsSidEntry struct {
	EnsaVersion string               `json:"ensa_version"`
	Instances   []AscsErsSidInstance `json:"instances"`
}

type AscsErsSidInstance struct {
	ResourceGroup    string `json:"resource_group"`
	ResourceInstance string `json:"resource_instance"`
	Name             string `json:"name"`
	InstanceNumber   string `json:"instance_number"`
	VirtualHostname  string `json:"virtual_hostname"`
	FilesystemBased  bool   `json:"filesystem_based"`
	Local            bool   `json:"local"`
}

type AscsErsClusterGatherer struct {
	executor   utils.CommandExecutor
	webService sapcontrolapi.WebServiceConnector
	cache      *factscache.FactsCache
}

func NewDefaultAscsErsClusterGatherer() *AscsErsClusterGatherer {
	webService := sapcontrolapi.WebServiceUnix{}
	return NewAscsErsClusterGatherer(utils.Executor{}, webService, nil)
}

func NewAscsErsClusterGatherer(executor utils.CommandExecutor, webService sapcontrolapi.WebServiceConnector,
	cache *factscache.FactsCache) *AscsErsClusterGatherer {
	return &AscsErsClusterGatherer{
		executor:   executor,
		webService: webService,
		cache:      cache,
	}
}

func (g *AscsErsClusterGatherer) SetCache(cache *factscache.FactsCache) {
	g.cache = cache
}

func (g *AscsErsClusterGatherer) Gather(
	_ context.Context,
	factsRequests []entities.FactRequest,
) ([]entities.Fact, error) {
	log.Infof("Starting %s facts gathering process", AscsErsClusterGathererName)
	var cibdata cib.Root

	ctx := context.Background()

	content, err := factscache.GetOrUpdate(
		g.cache,
		CibAdminGathererCache,
		makeMemoizeCibAdmin(ctx),
		g.executor,
	)

	if err != nil {
		return nil, CibAdminCommandError.Wrap(err.Error())
	}

	cibadmin, ok := content.([]byte)
	if !ok {
		return nil, AscsErsClusterDecodingError.Wrap("error casting the command output")
	}

	err = xml.Unmarshal(cibadmin, &cibdata)
	if err != nil {
		return nil, err
	}

	entries, err := getMultiSidEntries(ctx, g.cache, cibdata, g.webService)
	if err != nil {
		return nil, AscsErsClusterCibError.Wrap(err.Error())
	}

	factValues, err := mapMultiSidEntriesToFactValue(entries)
	if err != nil {
		return nil, AscsErsClusterDecodingError.Wrap(err.Error())
	}

	facts := []entities.Fact{}

	for _, requestedFact := range factsRequests {
		facts = append(facts, entities.NewFactGatheredWithRequest(requestedFact, factValues))
	}

	log.Infof("Requested %s facts gathered", AscsErsClusterGathererName)
	return facts, nil
}

func getMultiSidEntries(
	ctx context.Context,
	cache *factscache.FactsCache,
	cibdata cib.Root,
	webService sapcontrolapi.WebServiceConnector,
) (map[string]AscsErsSidEntry, error) {

	result := make(map[string]AscsErsSidEntry)

	for _, group := range cibdata.Configuration.Resources.Groups {
		var instanceGroupFound, filesystemBased bool
		var resourceGroup, resourceInstance, sid, instanceName, instanceNumber, virtualHostname string

		for _, primitive := range group.Primitives {
			for _, instanceAttribute := range primitive.InstanceAttributes {
				if instanceAttribute.Name == "InstanceName" {
					values := strings.Split(instanceAttribute.Value, "_")
					if len(values) != 3 {
						return nil, fmt.Errorf("incorrect InstanceName property value: %s", instanceAttribute.Value)
					}

					instanceGroupFound = true
					resourceGroup = group.ID
					resourceInstance = primitive.ID
					sid = values[0]
					instanceName = values[1]
					virtualHostname = values[2]
					if len(instanceName) < 2 {
						return nil, fmt.Errorf("incorrect instance name within the InstanceName value: %s", instanceName)
					}

					instanceNumber = instanceName[len(instanceName)-2:]
				}
			}

			if primitive.Type == "Filesystem" {
				filesystemBased = true
			}
		}

		if !instanceGroupFound {
			continue
		}

		ensaVersion, local := getEnsaVersionInfo(ctx, cache, webService, sid, instanceNumber)
		instance := AscsErsSidInstance{
			ResourceGroup:    resourceGroup,
			ResourceInstance: resourceInstance,
			Name:             instanceName,
			InstanceNumber:   instanceNumber,
			VirtualHostname:  virtualHostname,
			FilesystemBased:  filesystemBased,
			Local:            local,
		}

		entry, found := result[sid]

		if !found {
			entry = AscsErsSidEntry{
				EnsaVersion: EnsaUnknown,
				Instances:   []AscsErsSidInstance{instance},
			}
		} else {
			entry.Instances = append(entry.Instances, instance)
		}

		if local {
			entry.EnsaVersion = ensaVersion
		}

		result[sid] = entry
	}

	return result, nil
}

// getEnsaVersionInfo returns the ensa version and if the sap instance is running locally in this host
func getEnsaVersionInfo(
	ctx context.Context,
	cache *factscache.FactsCache,
	webService sapcontrolapi.WebServiceConnector,
	sid string,
	instanceNumber string,
) (string, bool) {

	cacheEntry := fmt.Sprintf("%s:%s:%s:%s", SapControlGathererCache, "GetProcessList", sid, instanceNumber)
	output, err := factscache.GetOrUpdate(
		cache,
		cacheEntry,
		memoizeSapcontrol,
		ctx,
		webService,
		instanceNumber,
		mapGetProcessList,
	)
	if err != nil {
		log.Warnf("error requesting GetProcessList information: %s", err.Error())
		return EnsaUnknown, false
	}

	processes, ok := output.([]*sapcontrolapi.OSProcess)
	if !ok {
		log.Warnf("error decoding GetProcessList information")
		return EnsaUnknown, false
	}

	for _, process := range processes {
		switch process.Name {
		case "enserver", "enrepserver":
			return Ensa1, true
		case "enq_server", "enq_replicator":
			return Ensa2, true
		}
	}

	return EnsaUnknown, true
}

func mapMultiSidEntriesToFactValue(entries map[string]AscsErsSidEntry) (entities.FactValue, error) {
	marshalled, err := json.Marshal(&entries)
	if err != nil {
		return nil, err
	}

	var unmarshalled map[string]interface{}
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return nil, err
	}

	return entities.NewFactValue(unmarshalled)
}
