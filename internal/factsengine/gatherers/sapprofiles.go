package gatherers

import (
	"encoding/json"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/trento-project/agent/internal/core/sapsystem"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	SapProfilesGathererName = "sap_profiles"
	sapMntPath              = "/sapmnt"
)

// nolint:gochecknoglobals
var (
	SapProfilesFileSystemError = entities.FactGatheringError{
		Type:    "sap-profiles-file-system-error",
		Message: "error reading the sap profiles file system",
	}

	SapProfilesDecodingError = entities.FactGatheringError{
		Type:    "sap-profiles-decoding-error",
		Message: "error deconding sap profiles content",
	}
)

type SapProfile struct {
	Name    string            `json:"name"`
	Path    string            `json:"path"`
	Content map[string]string `json:"content"`
}

type SapSystemEntry struct {
	Profiles []SapProfile `json:"profiles"`
}

type SapSystemMap map[string]SapSystemEntry

type SapProfilesGatherer struct {
	fs afero.Fs
}

func NewDefaultSapProfilesGatherer() *SapProfilesGatherer {
	return NewSapProfilesGatherer(afero.NewOsFs())
}

func NewSapProfilesGatherer(fs afero.Fs) *SapProfilesGatherer {
	return &SapProfilesGatherer{fs: fs}
}

func (s *SapProfilesGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	log.Infof("Starting %s facts gathering process", SapProfilesGathererName)
	facts := []entities.Fact{}
	systems := make(SapSystemMap)

	systemPaths, err := sapsystem.FindSystems(s.fs)
	if err != nil {
		return nil, SapProfilesFileSystemError.Wrap(err.Error())
	}

	for _, systemPath := range systemPaths {
		sid := filepath.Base(systemPath)
		profiles, err := mapSapProfiles(s.fs, sid)
		if err != nil {
			return nil, SapProfilesFileSystemError.Wrap(err.Error())
		}

		systems[sid] = SapSystemEntry{
			Profiles: profiles,
		}

	}

	factValues, err := systemsToFactValue(systems)
	if err != nil {
		return nil, SapProfilesDecodingError.Wrap(err.Error())
	}

	for _, requestedFact := range factsRequests {
		facts = append(facts, entities.NewFactGatheredWithRequest(requestedFact, factValues))
	}

	log.Infof("Requested %s facts gathered", SapProfilesGathererName)

	return facts, nil
}

func mapSapProfiles(fs afero.Fs, sid string) ([]SapProfile, error) {
	profiles := []SapProfile{}
	profileNames, err := sapsystem.FindProfiles(fs, sid)
	if err != nil {
		return nil, err
	}

	for _, profileName := range profileNames {
		profilePath := path.Join(sapMntPath, sid, "profile", profileName)
		content, err := sapsystem.GetProfileData(fs, profilePath)
		if err != nil {
			return nil, err
		}

		profiles = append(profiles, SapProfile{Name: profileName, Path: profilePath, Content: content})
	}

	return profiles, nil
}

func systemsToFactValue(profiles SapSystemMap) (entities.FactValue, error) {
	marshalled, err := json.Marshal(&profiles)
	if err != nil {
		return nil, err
	}

	var unmarshalled map[string]interface{}
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return nil, err
	}

	return entities.NewFactValue(unmarshalled, entities.WithStringConversion())
}
