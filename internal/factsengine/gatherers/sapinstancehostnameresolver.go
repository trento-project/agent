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
	SapInstanceHostnameResolverGathererName = "sapinstance_hostname_resolver"
)

// nolint:gochecknoglobals
var (
	hostnameRegexCompiled                   = regexp.MustCompile(`(.+)_(.+)_(.+)`) // <SID>_<InstanceNumber>_<Hostname>
	regexSubgroupsCount                     = 4
	SapInstanceHostnameResolverDetailsError = entities.FactGatheringError{
		Type:    "sapinstance-hostname-resolver-details-error",
		Message: "error gathering details",
	}
	SapInstanceHostnameResolverGathererDecodingError = entities.FactGatheringError{
		Type:    "sapinstance-hostname-resolver-decoding-error",
		Message: "error decoding output to FactValue",
	}
)

//go:generate mockery --name=HostnameResolver
type HostnameResolver interface {
	LookupHost(host string) ([]string, error)
}

//go:generate mockery --name=HostPinger
type HostPinger interface {
	Ping(host string) bool
}

type SapInstanceHostnameResolverGatherer struct {
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

func (p *Pinger) Ping(host string) bool {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return false
	}
	pinger.Count = 1
	err = pinger.Run()
	return err == nil
}

func NewDefaultSapInstanceHostnameResolverGatherer() *SapInstanceHostnameResolverGatherer {
	return NewSapInstanceHostnameResolverGatherer(afero.NewOsFs(), &Resolver{}, &Pinger{})
}

func NewSapInstanceHostnameResolverGatherer(
	fs afero.Fs,
	hr HostnameResolver,
	hp HostPinger) *SapInstanceHostnameResolverGatherer {

	return &SapInstanceHostnameResolverGatherer{fs: fs, hr: hr, hp: hp}
}

func (r *SapInstanceHostnameResolverGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}

	details, err := r.getInstanceHostnameDetails()
	if err != nil {
		log.Error(err)
		return nil, SapInstanceHostnameResolverDetailsError.Wrap(err.Error())
	}

	var fact entities.Fact
	factValue, err := mapReachabilityDetailsToFactValue(details)
	if err != nil {
		log.Error(err)
		return facts, &SapInstanceHostnameResolverGathererDecodingError
	}

	for _, factReq := range factsRequests {
		fact = entities.NewFactGatheredWithRequest(factReq, factValue)
		facts = append(facts, fact)
	}

	return facts, nil
}

func (r *SapInstanceHostnameResolverGatherer) getInstanceHostnameDetails() (map[string][]ResolvabilityDetails, error) {
	systems, err := sapsystem.FindSystems(r.fs)
	if err != nil {
		return nil, err
	}

	resolvabilityDetails := make(map[string][]ResolvabilityDetails)
	for _, system := range systems {
		sid := filepath.Base(system)
		profileFiles, err := sapsystem.FindProfiles(r.fs, sid)
		if err != nil {
			return nil, err
		}

		for _, profileFile := range profileFiles {
			if profileFile == sapsystem.SapDefaultProfile {
				continue
			}

			match := hostnameRegexCompiled.FindStringSubmatch(profileFile)
			if len(match) != regexSubgroupsCount {
				log.Error("error extracting SID/InstanceName/Hostname from profile file: ", profileFile)
				continue
			}
			matchedSID := match[1]
			matchedInstanceName := match[2]
			matchedHostname := match[3]

			addresses, err := r.hr.LookupHost(matchedHostname)
			if err != nil {
				log.Error("error resolving hostname: ", matchedHostname)
			}

			details := ResolvabilityDetails{
				Hostname:     matchedHostname,
				Addresses:    addresses,
				InstanceName: matchedInstanceName,
				Reachability: r.hp.Ping(matchedHostname),
			}

			if _, found := resolvabilityDetails[matchedSID]; !found {
				resolvabilityDetails[matchedSID] = []ResolvabilityDetails{details}
			} else {
				resolvabilityDetails[matchedSID] = append(resolvabilityDetails[matchedSID], details)
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
