package gatherers

import (
	"encoding/json"
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
	hostnameParsingRegexp            = `(?P<SID>[A-Z0-9]+)_(?P<InstanceName>[A-Z0-9]+)_(?P<Hostname>[a-z]+)$`
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
)

// saplocalhost_reachability
type SapLocalhostResolverGatherer struct {
	fs afero.Fs
	hr utils.HostnameResolver
}

type ReachabilityDetails struct {
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
		factValue, _ := mapReachabilityDetailsToFactValue(rd)
		fact = entities.NewFactGatheredWithRequest(factReq, factValue)
		facts = append(facts, fact)
	}

	return facts, nil
}

func (r *SapLocalhostResolverGatherer) getInstanceHostnameDetails() (map[string]ReachabilityDetails, error) {
	systems, err := sapsystem.FindSystems(r.fs)
	reachabilityDetails := make(map[string]ReachabilityDetails)

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
			result := make(map[string]string)
			match := hostnameRegexCompiled.FindStringSubmatch(profileFile)
			if len(match) == 0 {
				continue
			}
			for i, name := range hostnameRegexCompiled.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = match[i]
				}
			}
			details := ReachabilityDetails{
				Hostname:     result["Hostname"],
				Addresses:    []string{},
				InstanceName: result["InstanceName"],
			}
			details.Addresses, _ = r.hr.LookupHost(details.Hostname)
			reachabilityDetails[result["SID"]] = details
		}
	}

	return reachabilityDetails, nil
}

func mapReachabilityDetailsToFactValue(entries map[string]ReachabilityDetails) (entities.FactValue, error) {
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
