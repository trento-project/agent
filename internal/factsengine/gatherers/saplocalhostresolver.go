package gatherers

import (
	"net"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/trento-project/agent/internal/core/sapsystem"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const SapLocalhostResolverGathererName = "saplocalhost_resolver"

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

func NewDefaultSapLocalhostResolverGatherer() *SapLocalhostResolverGatherer {
	return NewSapLocalhostResolver(afero.NewOsFs(), utils.Resolver{})
}

func NewSapLocalhostResolver(fs afero.Fs, hr utils.HostnameResolver) *SapLocalhostResolverGatherer {
	return &SapLocalhostResolverGatherer{fs: fs, hr: hr}
}

func (r *SapLocalhostResolverGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	var facts []entities.Fact
	profiles, _ := r.getProfiles()

	for _, factReq := range factsRequests {
		var fact entities.Fact
		for _, profile := range profiles {
			ips, err := r.hr.LookupHost(profile["SAPLOCALHOST"])
			if err != nil {
				log.Error(err)
				gatheringError := SapLocalhostResolverHostnameResolutionError.Wrap(profile["SAPLOCALHOST"])
				fact = entities.NewFactGatheredWithError(factReq, gatheringError)
			} else {
				ipList := &entities.FactValueList{}
				for _, ip := range ips {
					ipList.Value = append(ipList.Value, &entities.FactValueString{Value: ip})
				}

				// reachable := false
				// for _, ip := range ips {
				// 	if ok, _ := isHostReachable(ip); ok {
				// 		reachable = true
				// 		break
				// 	}
				// }

				factValue := &entities.FactValueMap{
					Value: map[string]entities.FactValue{
						profile["SAPSYSTEMNAME"]: &entities.FactValueList{
							Value: []entities.FactValue{
								&entities.FactValueMap{
									Value: map[string]entities.FactValue{
										"hostname":      &entities.FactValueString{Value: profile["SAPLOCALHOST"]},
										"addresses":     ipList,
										"instance_name": &entities.FactValueString{Value: profile["INSTANCE_NAME"]},
										// "reachable":     &entities.FactValueBool{Value: reachable},
									},
								},
							},
						},
					},
				}
				fact = entities.NewFactGatheredWithRequest(factReq, factValue)
			}
			facts = append(facts, fact)
		}
	}

	return facts, nil
}

func isHostReachable(ip string) (bool, error) {
	conn, err := net.DialTimeout("ip4:icmp", ip, time.Second)
	if err != nil {
		log.Error(err)
		return false, err
	}
	defer conn.Close()

	return true, nil
}

func (r *SapLocalhostResolverGatherer) getSapSystemInstances() ([]string, error) {
	systems, err := sapsystem.FindSystems(r.fs)

	if err != nil {
		return nil, err
	}

	sysNames := GetSIDsString(systems)
	return sysNames, nil
}

func (r *SapLocalhostResolverGatherer) getProfiles() ([]map[string]string, error) {
	sids, _ := r.getSapSystemInstances()
	var profiles []map[string]string

	for _, sid := range sids {
		profileNames, err := sapsystem.FindProfiles(r.fs, sid)
		if err != nil {
			return nil, err
		}

		for _, profile := range profileNames {
			if profile == "DEFAULT.PFL" {
				continue
			}
			profilePath := filepath.Join(sapMntPath, sid, "profile", profile)
			content, err := sapsystem.GetProfileData(r.fs, profilePath)
			if err != nil {
				return nil, err
			}

			if _, ok := content["SAPLOCALHOST"]; ok {
				profiles = append(profiles, content)
			}
		}
	}
	return profiles, nil
}

func GetSIDsString(paths []string) []string {
	names := make([]string, len(paths))
	for i, path := range paths {
		names[i] = filepath.Base(path)
	}
	return names
}
