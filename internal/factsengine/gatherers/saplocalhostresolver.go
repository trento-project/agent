package gatherers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/trento-project/agent/internal/core/sapsystem"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	SapLocalhostResolverGathererName = "saplocalhost_resolver"
)

var subgroupMapping = map[int]string{
	1: "SID",
	2: "InstanceName",
	3: "Hostname",
}

var hostnameParsingRegexp = fmt.Sprintf(
	`(?P<%s>[A-Z0-9]+)_(?P<%s>[A-Z0-9]+)_(?P<%s>[a-z]+)$`,
	subgroupMapping[1],
	subgroupMapping[2],
	subgroupMapping[3],
)

var (
	hostnameRegexCompiled = regexp.MustCompile(hostnameParsingRegexp)
)

// nolint:gochecknoglobals
var (
	SapLocalhostResolverHostnameResolutionError = entities.FactGatheringError{
		Type:    "saplocalhost_resolver-resolution-error",
		Message: "error resolving hostname",
	}
	SapLocalhostResolverGathererDecodingError = entities.FactGatheringError{
		Type:    "saplocalhost_resolver-decoding-error",
		Message: "error decoding output to FactValue",
	}
)

type SapLocalhostResolverGatherer struct {
	fs afero.Fs
	hr utils.HostnameResolver
}

type ResolvabilityDetails struct {
	Hostname     string   `json:"hostname"`
	Addresses    []string `json:"addresses"`
	InstanceName string   `json:"instance_name"`
}

func NewDefaultSapLocalhostResolverGatherer() *SapLocalhostResolverGatherer {
	return NewSapLocalhostResolver(afero.NewOsFs(), utils.Resolver{})
}

func NewSapLocalhostResolver(fs afero.Fs, hr utils.HostnameResolver) *SapLocalhostResolverGatherer {
	return &SapLocalhostResolverGatherer{fs: fs, hr: hr}
}

func (r *SapLocalhostResolverGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	var facts []entities.Fact

	rd, err := r.getInstanceHostnameDetails()
	if err != nil {
		log.Error(err)
		return nil, SapLocalhostResolverHostnameResolutionError.Wrap(err.Error())
	}

	for _, factReq := range factsRequests {
		var fact entities.Fact
		factValue, err := mapReachabilityDetailsToFactValue(rd)
		if err != nil {
			log.Error(err)
			fact = entities.NewFactGatheredWithError(factReq, SapLocalhostResolverGathererDecodingError.Wrap(err.Error()))
		} else {
			fact = entities.NewFactGatheredWithRequest(factReq, factValue)
		}
		facts = append(facts, fact)
	}

	return facts, nil
}

func (r *SapLocalhostResolverGatherer) getInstanceHostnameDetails() (map[string]ResolvabilityDetails, error) {
	systems, err := sapsystem.FindSystems(r.fs)
	reachabilityDetails := make(map[string]ResolvabilityDetails)

	if err != nil {
		return nil, err
	}

	for _, system := range systems {
		sid := filepath.Base(system)
		profileFiles, err := sapsystem.FindProfiles(r.fs, sid)
		if err != nil {
			return nil, err
		}

		for _, profileFile := range profileFiles {
			if filepath.Base(profileFile) == "DEFAULT.PFL" {
				continue
			}

			match := hostnameRegexCompiled.FindStringSubmatch(profileFile)
			if len(match) != len(subgroupMapping)+1 {
				continue
			}
			addresses, err := r.hr.LookupHost(match[3])
			if err != nil {
				return nil, err
			}

			details := ResolvabilityDetails{
				Hostname:     match[3],
				Addresses:    addresses,
				InstanceName: match[2],
			}
			reachabilityDetails[match[1]] = details
		}
	}

	return reachabilityDetails, nil
}

func mapReachabilityDetailsToFactValue(entries map[string]ResolvabilityDetails) (entities.FactValue, error) {
	marshalled, err := json.Marshal(&entries)
	if err != nil {
		return nil, err
	}

	var unmarshalled interface{}
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return nil, err
	}

	return entities.NewFactValue(unmarshalled)
}
