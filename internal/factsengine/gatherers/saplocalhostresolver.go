package gatherers

import (
	"encoding/json"
	"net"
	"path/filepath"
	"regexp"

	probing "github.com/prometheus-community/pro-bing"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/trento-project/agent/internal/core/sapsystem"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	SapLocalhostResolverGathererName = "saplocalhost_resolver"
)

// nolint:gochecknoglobals
var (
	hostnameRegexCompiled = regexp.MustCompile(`(.+)_(.+)_(.+)`) //<SID>_<InstanceNumber>_<Hostname>
	regexSubgroupsCount   = 4
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
	SapLocalhostResolverFileSystemError = entities.FactGatheringError{
		Type:    "saplocalhost_resolver-file-system-error",
		Message: "error reading the sap profiles file system",
	}
)

//go:generate mockery --name=HostnameResolver
type HostnameResolver interface {
	LookupHost(host string) ([]string, error)
}

//go:generate mockery --name=HostPinger
type HostPinger interface {
	Ping(name string, arg ...string) bool
}

type SapLocalhostResolverGatherer struct {
	fs afero.Fs
	hr HostnameResolver
	hp HostPinger
}

type ResolvabilityDetails struct {
	Hostname     string   `json:"hostname"`
	Addresses    []string `json:"addresses"`
	InstanceName string   `json:"instance_name"`
	Reachability bool     `json:"reachability"`
}

type Resolver struct{}

type Pinger struct{}

func (r *Resolver) LookupHost(host string) ([]string, error) {
	return net.LookupHost(host)
}

func (p Pinger) Ping(name string, arg ...string) bool {
	pinger, err := probing.NewPinger(name)
	if err != nil {
		return false
	}
	pinger.Count = 3
	err = pinger.Run()
	return err == nil
}

func NewDefaultSapLocalhostResolverGatherer() *SapLocalhostResolverGatherer {
	return NewSapLocalhostResolver(afero.NewOsFs(), &Resolver{}, Pinger{})
}

func NewSapLocalhostResolver(fs afero.Fs, hr HostnameResolver, hp HostPinger) *SapLocalhostResolverGatherer {
	return &SapLocalhostResolverGatherer{fs: fs, hr: hr, hp: hp}
}

func (r *SapLocalhostResolverGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := make([]entities.Fact, 0)

	details, err := r.getInstanceHostnameDetails()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	for _, factReq := range factsRequests {
		var fact entities.Fact
		factValue, err := mapReachabilityDetailsToFactValue(details)
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

func (r *SapLocalhostResolverGatherer) getInstanceHostnameDetails() (map[string][]ResolvabilityDetails, error) {
	systems, err := sapsystem.FindSystems(r.fs)
	if err != nil {
		return nil, err
	}

	resolvabilityDetails := make(map[string][]ResolvabilityDetails)
	for _, system := range systems {
		sid := filepath.Base(system)
		profileFiles, err := sapsystem.FindProfiles(r.fs, sid)
		if err != nil {
			return nil, SapLocalhostResolverFileSystemError.Wrap(err.Error())
		}

		for _, profileFile := range profileFiles {
			if profileFile == sapsystem.SapDefaultProfile {
				continue
			}

			match := hostnameRegexCompiled.FindStringSubmatch(profileFile)
			if len(match) != regexSubgroupsCount {
				continue
			}
			matchedSID := match[1]
			matchedInstanceName := match[2]
			matchedHostname := match[3]

			addresses, err := r.hr.LookupHost(matchedHostname)
			if err != nil {
				return nil, SapLocalhostResolverHostnameResolutionError.Wrap(err.Error())
			}

			details := ResolvabilityDetails{
				Hostname:     matchedHostname,
				Addresses:    addresses,
				InstanceName: matchedInstanceName,
				Reachability: r.hp.Ping(matchedHostname),
			}
			if _, found := resolvabilityDetails[match[1]]; !found {
				resolvabilityDetails[matchedSID] = []ResolvabilityDetails{details}
			} else {
				resolvabilityDetails[matchedSID] = append(resolvabilityDetails[match[1]], details)
			}
		}
	}

	return resolvabilityDetails, nil
}

func mapReachabilityDetailsToFactValue(entries map[string][]ResolvabilityDetails) (entities.FactValue, error) {
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
