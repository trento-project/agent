package gatherers

import (
	"encoding/json"
	"io/fs"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/trento-project/agent/internal/core/sapsystem"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	SapProfilesGathererName = "sap_profiles"
	sapFolder               = "/sapmnt"
	profileFilePattern      = "^/sapmnt/([A-Z][A-Z0-9]{2})/profile/(DEFAULT\\.PFL|[^.]*)$"
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
	SID     string            `json:"sid"`
	Name    string            `json:"name"`
	Path    string            `json:"path"`
	Content map[string]string `json:"content"`
}

type SapProfilesGatherer struct {
	fs      afero.Fs
	pattern *regexp.Regexp
}

func NewDefaultSapProfilesGatherer() *SapProfilesGatherer {
	return NewSapProfilesGatherer(afero.NewOsFs())
}

func NewSapProfilesGatherer(fs afero.Fs) *SapProfilesGatherer {
	pattern := regexp.MustCompile(profileFilePattern)
	return &SapProfilesGatherer{fs: fs, pattern: pattern}
}

func (s *SapProfilesGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	log.Infof("Starting %s facts gathering process", SapProfilesGathererName)
	facts := []entities.Fact{}
	sapProfiles := []SapProfile{}

	err := afero.Walk(s.fs, sapFolder, func(filePath string, info fs.FileInfo, err error) error {
		matches := s.pattern.FindStringSubmatch(filePath)
		if matches != nil {
			sid := matches[1]
			content, err := sapsystem.GetProfileData(s.fs, filePath)
			if err != nil {
				return err
			}
			sapProfiles = append(sapProfiles, SapProfile{SID: sid, Name: info.Name(), Path: filePath, Content: content})
		}

		return nil
	})

	if err != nil {
		return nil, SapProfilesFileSystemError.Wrap(err.Error())
	}

	factValues, err := profilesToFactValue(sapProfiles)
	if err != nil {
		return nil, SapProfilesDecodingError.Wrap(err.Error())
	}

	for _, requestedFact := range factsRequests {
		facts = append(facts, entities.NewFactGatheredWithRequest(requestedFact, factValues))
	}

	log.Infof("Requested %s facts gathered", SapProfilesGathererName)

	return facts, nil
}

func profilesToFactValue(profiles []SapProfile) (entities.FactValue, error) {
	marshalled, err := json.Marshal(&profiles)
	if err != nil {
		return nil, err
	}

	var unmarshalled []interface{}
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return nil, err
	}

	return entities.NewFactValue(unmarshalled, entities.WithStringConversion())
}
